package chatmodel

import (
	"context"
	"fmt"
	"log"

	"gogogajeto/agent/prompts"
	"gogogajeto/util"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type State struct {
	History   []*schema.Message
	UserInput string
	Name      string
}

const (
	NodeKeyHuman         = "Human"
	NodeKeyInputConvert  = "InputConverter"
	NodeKeyChatModel     = "ChatModel"
	NodeKeyToolsNode     = "ToolsNode"
	NodeKeyOutputConvert = "OutputConverter"
)

func ComposeAgent(ctx context.Context,
	cm model.BaseChatModel,
	tools []tool.BaseTool,
) compose.Runnable[string, string] {
	g := compose.NewGraph[string, string](compose.WithGenLocalState(func(ctx context.Context) *State {
		return &State{History: []*schema.Message{}}
	}))

	// create and register nodes with logging callbacks
	err := g.AddLambdaNode(NodeKeyInputConvert, compose.InvokableLambda(func(ctx context.Context, input string) (output []*schema.Message, err error) {
		util.LogMessage("=== InputConvert Node START ===")
		util.LogMessage("Input: " + input)

		messages := []*schema.Message{
			schema.SystemMessage(prompts.SystemPrompt),
			schema.UserMessage(input),
		}

		util.LogMessage(fmt.Sprintf("Generated messages count: %d", len(messages)))
		util.LogMessage("=== InputConvert Node END ===")

		return messages, nil
	}), compose.WithNodeName(NodeKeyInputConvert))
	if err != nil {
		log.Fatal(err)
	}

	err = g.AddChatModelNode(
		NodeKeyChatModel,
		cm,
		compose.WithNodeName(NodeKeyChatModel),
		// Pre-handler logging
		compose.WithStatePreHandler(func(ctx context.Context, in []*schema.Message, state *State) ([]*schema.Message, error) {
			util.LogMessage("=== ChatModel Node PRE-HANDLER ===")
			util.LogMessage(fmt.Sprintf("Input messages count: %d", len(in)))
			util.LogMessage(fmt.Sprintf("Current history count: %d", len(state.History)))

			state.History = append(state.History, in...)

			util.LogMessage(fmt.Sprintf("Updated history count: %d", len(state.History)))

			// Log the last few messages in history to debug tool call/response matching
			historyLen := len(state.History)
			startIdx := historyLen - 5
			if startIdx < 0 {
				startIdx = 0
			}

			util.LogMessage("=== Recent History for OpenAI ===")
			for i := startIdx; i < historyLen; i++ {
				msg := state.History[i]
				util.LogMessage(fmt.Sprintf("History[%d] Role: %s", i, string(msg.Role)))
				if msg.Role == schema.Assistant && len(msg.ToolCalls) > 0 {
					util.LogMessage(fmt.Sprintf("History[%d] has %d tool calls", i, len(msg.ToolCalls)))
					for j, tc := range msg.ToolCalls {
						util.LogMessage(fmt.Sprintf("  ToolCall[%d] ID: %s, Name: %s", j, tc.ID, tc.Function.Name))
					}
				}
				if msg.Role == schema.Tool {
					util.LogMessage(fmt.Sprintf("History[%d] ToolCallID: %s, Name: %s", i, msg.ToolCallID, msg.Name))
				}
			}
			util.LogMessage("=== End Recent History ===")

			return state.History, nil
		}),
		// Post-handler logging
		compose.WithStatePostHandler(func(ctx context.Context, out *schema.Message, state *State) (*schema.Message, error) {
			util.LogMessage("=== ChatModel Node POST-HANDLER ===")
			util.LogMessage("Output message role: " + string(out.Role))
			util.LogMessage(fmt.Sprintf("Output content length: %d", len(out.Content)))
			util.LogMessage(fmt.Sprintf("Tool calls count: %d", len(out.ToolCalls)))

			// Log tool calls if any
			if len(out.ToolCalls) > 0 {
				for i, toolCall := range out.ToolCalls {
					util.LogMessage(fmt.Sprintf("Tool call %d: %s", i, toolCall.Function.Name))
					util.LogMessage(fmt.Sprintf("Tool call %d ID: %s", i, toolCall.ID))
					util.LogMessage("Tool arguments: " + toolCall.Function.Arguments)
				}
			}

			state.History = append(state.History, out)
			util.LogMessage("=== ChatModel Node END ===")

			return out, nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{Tools: append(tools)})
	if err != nil {
		log.Fatal(err)
	}
	err = g.AddToolsNode(
		NodeKeyToolsNode,
		toolsNode,
		compose.WithNodeName(NodeKeyToolsNode),
		// Pre-handler for tools
		compose.WithStatePreHandler(func(ctx context.Context, in *schema.Message, state *State) (*schema.Message, error) {
			util.LogMessage("=== ToolsNode PRE-HANDLER ===")
			util.LogMessage("Input message role: " + string(in.Role))

			if len(in.ToolCalls) > 0 {
				util.LogMessage(fmt.Sprintf("Tools to execute: %d", len(in.ToolCalls)))
				for i, toolCall := range in.ToolCalls {
					util.LogMessage(fmt.Sprintf("Executing tool %d: %s", i, toolCall.Function.Name))
					util.LogMessage("With arguments: " + toolCall.Function.Arguments)
				}
			}

			return in, nil
		}),
		// Post-handler for tools
		compose.WithStatePostHandler(func(ctx context.Context, out []*schema.Message, state *State) ([]*schema.Message, error) {
			util.LogMessage("=== ToolsNode POST-HANDLER ===")
			util.LogMessage(fmt.Sprintf("Tool execution results count: %d", len(out)))

			for i, msg := range out {
				util.LogMessage(fmt.Sprintf("Message %d role: %s", i, string(msg.Role)))
				if msg.Role == schema.Tool {
					util.LogMessage(fmt.Sprintf("Tool result %d content: %s", i, msg.Content))
					util.LogMessage(fmt.Sprintf("Tool result %d ToolCallID: %s", i, msg.ToolCallID))
					util.LogMessage(fmt.Sprintf("Tool result %d Name: %s", i, msg.Name))
				}
			}

			// DO NOT append to state.History here - ChatModel pre-handler will handle it
			// state.History = append(state.History, out...)  // ← REMOVED to prevent duplicate
			util.LogMessage("=== ToolsNode END ===")

			return out, nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = g.AddLambdaNode(NodeKeyHuman, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (output []*schema.Message, err error) {
		util.LogMessage("=== Human Node START ===")
		util.LogMessage("Input message role: " + string(input.Role))
		util.LogMessage("Input content: " + input.Content)

		return []*schema.Message{input}, nil
	}), compose.WithNodeName(NodeKeyHuman),
		compose.WithStatePostHandler(func(ctx context.Context, in []*schema.Message, state *State) ([]*schema.Message, error) {
			util.LogMessage("=== Human Node POST-HANDLER ===")
			util.LogMessage("UserInput from state: " + state.UserInput)

			if len(state.UserInput) > 0 {
				userMsg := schema.UserMessage(state.UserInput)
				util.LogMessage("Creating new user message: " + userMsg.Content)
				util.LogMessage("=== Human Node END ===")
				return []*schema.Message{userMsg}, nil
			}
			util.LogMessage("=== Human Node END ===")
			return in, nil
		}))
	if err != nil {
		log.Fatal(err)
	}

	err = g.AddLambdaNode(NodeKeyOutputConvert, compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) (output string, err error) {
		util.LogMessage("=== OutputConvert Node START ===")
		util.LogMessage(fmt.Sprintf("Input messages count: %d", len(input)))

		if len(input) > 0 {
			finalContent := input[len(input)-1].Content
			util.LogMessage(fmt.Sprintf("Final output content length: %d", len(finalContent)))
			util.LogMessage("=== OutputConvert Node END ===")
			return finalContent, nil
		}

		util.LogMessage("No messages to convert")
		util.LogMessage("=== OutputConvert Node END ===")
		return "", nil
	}))
	if err != nil {
		log.Fatal(err)
	}

	// compose graph
	err = g.AddEdge(compose.START, NodeKeyInputConvert)
	if err != nil {
		log.Fatal(err)
	}
	err = g.AddEdge(NodeKeyInputConvert, NodeKeyChatModel)
	if err != nil {
		log.Fatal(err)
	}
	err = g.AddBranch(NodeKeyChatModel, compose.NewGraphBranch(func(ctx context.Context, in *schema.Message) (endNode string, err error) {
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
	err = g.AddBranch(NodeKeyHuman, compose.NewGraphBranch(func(ctx context.Context, in []*schema.Message) (endNode string, err error) {
		if in[len(in)-1].Role == schema.User {
			return NodeKeyChatModel, nil
		}
		return NodeKeyOutputConvert, nil
	}, map[string]bool{
		NodeKeyChatModel:     true,
		NodeKeyOutputConvert: true,
	}))
	err = g.AddEdge(NodeKeyToolsNode, NodeKeyChatModel)
	if err != nil {
		log.Fatal(err)
	}
	err = g.AddEdge(NodeKeyOutputConvert, compose.END)
	if err != nil {
		log.Fatal(err)
	}

	runner, err := g.Compile(ctx, compose.WithCheckPointStore(NewInMemoryStore()), compose.WithInterruptBeforeNodes([]string{NodeKeyHuman}))
	if err != nil {
		log.Fatal(err)
	}

	return runner
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{m: make(map[string][]byte)}
}

type InMemoryStore struct {
	m map[string][]byte
}

func (i *InMemoryStore) Get(ctx context.Context, checkPointID string) ([]byte, bool, error) {
	data, ok := i.m[checkPointID]
	return data, ok, nil
}

func (i *InMemoryStore) Set(ctx context.Context, checkPointID string, checkPoint []byte) error {
	i.m[checkPointID] = checkPoint
	return nil
}
