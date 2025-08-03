package main

import (
	"context"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
	"log"
	"os"
	"sync"
)

var (
	openaiAPIKey  string
	openaiBaseURL string
	openaiModel   string

	input string
)

type state struct {
	History   []*schema.Message
	UserInput string
}

const (
	NodeKeyHuman         = "Human"
	NodeKeyInputConvert  = "InputConverter"
	NodeKeyChatModel     = "ChatModel"
	NodeKeyToolsNode     = "ToolsNode"
	NodeKeyOutputConvert = "OutputConverter"
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found or could not be loaded.")
	}

	openaiAPIKey = os.Getenv("OPENAI_API_KEY")
	openaiModel = os.Getenv("OPENAI_MODEL")
	openaiBaseURL = os.Getenv("OPENAI_API_BASE")

	if openaiAPIKey == "" || openaiModel == "" || openaiBaseURL == "" {
		log.Fatal("Error: Required environment variables (OPENAI_API_KEY, OPENAI_MODEL, OPENAI_API_BASE) are not set.")
	}

	input = "what is eino?"
}

type ChatModel struct {
}

func createAgent() compose.Runnable[string, string] {
	ctx := context.Background()

	// init sandbox tool and commandline tool
	sb := newSandbox(ctx)
	defer sb.Cleanup(ctx)
	commandlineTool := newCommandLineTool(ctx, sb)

	// init chat model and bind tools
	cm := newChatModel(ctx)
	cm = bindTools(ctx, cm, append(commandlineTool))

	// create agent
	agent := composeAgent(ctx, cm, commandlineTool)

	return agent
}

func newChatModel(ctx context.Context) model.ToolCallingChatModel {
	var cm model.ToolCallingChatModel
	var err error
	var temp float32 = 0
	logMessage(openaiAPIKey)
	cm, err = openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:      openaiAPIKey,
		BaseURL:     openaiBaseURL,
		Model:       openaiModel,
		Temperature: &temp,
		ByAzure:     false,
	})

	if err != nil {
		log.Fatalf("Failed to create OpenAI ChatModel: %v", err)
	}
	return cm
}

func composeAgent(ctx context.Context,
	cm model.BaseChatModel,
	tools []tool.BaseTool,
) compose.Runnable[string, string] {
	// Only register types once (singleton pattern)
	registerSerializableTypesOnce()

	g := compose.NewGraph[string, string](compose.WithGenLocalState(func(ctx context.Context) *state {
		return &state{History: []*schema.Message{}}
	}))

	// create nodes
	addInputConvertNode(g)
	addEdgeStartToInputConvert(g)
	addEdgeInputConvertToChatModel(g)

	addChatModelNode(g, cm)
	addBranchChatModel(g)
	
	addHumanNode(g)
	addBranchHuman(g)
	addToolsNode(g, tools) // <-- Add this line to ensure ToolsNode is added
	addOutputConvertNode(g)
	addEdgeOutputConvertToEnd(g)

	// create edges and branches

	runner, err := g.Compile(ctx, compose.WithCheckPointStore(newInMemoryStore()), compose.WithInterruptBeforeNodes([]string{NodeKeyHuman}))
	if err != nil {
		log.Fatal(err)
	}

	return runner
}

// Helper functions for node creation
func addInputConvertNode(g *compose.Graph[string, string]) {
	err := g.AddLambdaNode(NodeKeyInputConvert, compose.InvokableLambda(func(ctx context.Context, input string) (output []*schema.Message, err error) {
		return []*schema.Message{
			schema.SystemMessage(systemPrompt),
			schema.UserMessage(input),
		}, nil
	}), compose.WithNodeName(NodeKeyInputConvert))
	if err != nil {
		log.Fatal(err)
	}
}

func addChatModelNode(g *compose.Graph[string, string], cm model.BaseChatModel) {
	err := g.AddChatModelNode(
		NodeKeyChatModel,
		cm,
		compose.WithNodeName(NodeKeyChatModel),
		compose.WithStatePreHandler(func(ctx context.Context, in []*schema.Message, state *state) ([]*schema.Message, error) {
			state.History = append(state.History, in...)
			return state.History, nil
		}),
		compose.WithStatePostHandler(func(ctx context.Context, out *schema.Message, state *state) (*schema.Message, error) {
			state.History = append(state.History, out)
			return out, nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func addHumanNode(g *compose.Graph[string, string]) {
	err := g.AddLambdaNode(NodeKeyHuman, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (output []*schema.Message, err error) {
		return []*schema.Message{input}, nil
	}), compose.WithNodeName(NodeKeyHuman),
		compose.WithStatePostHandler(func(ctx context.Context, in []*schema.Message, state *state) ([]*schema.Message, error) {
			if len(state.UserInput) > 0 {
				return []*schema.Message{schema.UserMessage(state.UserInput)}, nil
			}
			return in, nil
		}))
	if err != nil {
		log.Fatal(err)
	}
}

func addToolsNode(g *compose.Graph[string, string], tools []tool.BaseTool) {
	if len(tools) == 0 {
		log.Println("Warning: No tools available for ToolsNode, using dummy logic.")
	}

	// Ensure tools are registered in the ToolsNode
	for _, t := range tools {
		info, err := t.Info(context.Background())
		if err != nil {
			log.Fatalf("Failed to get tool info: %v", err)
		}
		log.Printf("Registered tool: %s", info.Name)
	}

	err := g.AddLambdaNode(NodeKeyToolsNode, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (output *schema.Message, err error) {
		return schema.SystemMessage("[Tool logic executed]"), nil
	}), compose.WithNodeName(NodeKeyToolsNode))
	if err != nil {
		log.Fatal("Failed to add ToolsNode to the graph: ", err)
	}
}

func addOutputConvertNode(g *compose.Graph[string, string]) {
	err := g.AddLambdaNode(NodeKeyOutputConvert, compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) (output string, err error) {
		return input[len(input)-1].Content, nil
	}))
	if err != nil {
		log.Fatal(err)
	}
}

// Helper functions for each individual edge/branch
func addEdgeStartToInputConvert(g *compose.Graph[string, string]) {
	err := g.AddEdge(compose.START, NodeKeyInputConvert)
	if err != nil {
		log.Fatal(err)
	}
}

func addEdgeInputConvertToChatModel(g *compose.Graph[string, string]) {
	err := g.AddEdge(NodeKeyInputConvert, NodeKeyChatModel)
	if err != nil {
		log.Fatal(err)
	}
}

func addBranchChatModel(g *compose.Graph[string, string]) {
	err := g.AddBranch(NodeKeyChatModel, compose.NewGraphBranch(func(ctx context.Context, in *schema.Message) (endNode string, err error) {
		if len(in.ToolCalls) > 0 {
			return NodeKeyToolsNode, nil
		}
		return NodeKeyHuman, nil
	}, map[string]bool{
		NodeKeyToolsNode: true,
		NodeKeyHuman:     true,
	}))
	if err != nil {
		log.Fatal(err)
	}
}

func addBranchHuman(g *compose.Graph[string, string]) {
	err := g.AddBranch(NodeKeyHuman, compose.NewGraphBranch(func(ctx context.Context, in []*schema.Message) (endNode string, err error) {
		if in[len(in)-1].Role == schema.User {
			return NodeKeyChatModel, nil
		}
		return NodeKeyOutputConvert, nil
	}, map[string]bool{
		NodeKeyChatModel:     true,
		NodeKeyOutputConvert: true,
	}))
	if err != nil {
		log.Fatal(err)
	}
}

func addEdgeOutputConvertToEnd(g *compose.Graph[string, string]) {
	err := g.AddEdge(NodeKeyOutputConvert, compose.END)
	if err != nil {
		log.Fatal(err)
	}
}

func newInMemoryStore() *inMemoryStore {
	return &inMemoryStore{m: make(map[string][]byte)}
}

type inMemoryStore struct {
	m map[string][]byte
}

func (i *inMemoryStore) Get(ctx context.Context, checkPointID string) ([]byte, bool, error) {
	data, ok := i.m[checkPointID]
	return data, ok, nil
}

func (i *inMemoryStore) Set(ctx context.Context, checkPointID string, checkPoint []byte) error {
	i.m[checkPointID] = checkPoint
	return nil
}

var registerTypesOnce sync.Once

func registerSerializableTypesOnce() {
	registerTypesOnce.Do(func() {
		_ = compose.RegisterSerializableType[state]("my state")
		_ = compose.RegisterSerializableType[schema.ChatMessagePartType]("cmpt")
		_ = compose.RegisterSerializableType[schema.ChatMessageImageURL]("cmiu")
		_ = compose.RegisterSerializableType[schema.ChatMessageAudioURL]("cnau")
		_ = compose.RegisterSerializableType[schema.ChatMessageVideoURL]("cmvu")
		_ = compose.RegisterSerializableType[schema.ChatMessageFileURL]("cmfu")
		_ = compose.RegisterSerializableType[schema.ImageURLDetail]("iud")
	})
}
