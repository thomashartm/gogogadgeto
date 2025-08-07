package tools

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"gogogajeto/util"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino-ext/components/tool/commandline/sandbox"
	"github.com/cloudwego/eino/components/tool"
	"github.com/joho/godotenv"
)

var (
	openaiAPIKey  string
	openaiBaseURL string
	openaiModel   string
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
}

func NewChatModel(ctx context.Context) model.ToolCallingChatModel {
	var cm model.ToolCallingChatModel
	var err error
	var temp float32 = 0
	util.LogMessage("OpenAI API Key: " + openaiAPIKey)
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

func NewSandbox(ctx context.Context) *sandbox.DockerSandbox {
	sb, err := sandbox.NewDockerSandbox(ctx, &sandbox.Config{
		Image:          "python:3.11-slim",
		HostName:       "sandbox",
		WorkDir:        "/workspace",
		MemoryLimit:    512 * 1024 * 1024,
		CPULimit:       1.0,
		NetworkEnabled: false,
		Timeout:        time.Second * 30,
	})
	if err != nil {
		log.Fatal(err)
	}
	err = sb.Create(ctx)
	if err != nil {
		log.Fatal(err)
	}
	return sb
}

// Dummy tool fallback
func NewDummyTool() tool.BaseTool {
	return &dummyTool{Name: "dummy"}
}

type dummyTool struct {
	Name string
}

func (d *dummyTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: d.Name,
		// Description field removed for compatibility
	}, nil
}

func (d *dummyTool) Call(ctx context.Context, input any) (any, error) {
	return "[Dummy tool called]", nil
}

func NewCommandLineTool(ctx context.Context, sb commandline.Operator) []tool.BaseTool {
	et, err := commandline.NewStrReplaceEditor(ctx, &commandline.EditorConfig{Operator: sb})
	if err != nil {
		log.Fatal(err)
	}
	pt, err := commandline.NewPyExecutor(ctx, &commandline.PyExecutorConfig{Command: "python3", Operator: sb})
	if err != nil {
		log.Fatal(err)
	}
	return []tool.BaseTool{et, pt}
}

func BindTools(ctx context.Context, cm model.ToolCallingChatModel, tools []tool.BaseTool) model.ToolCallingChatModel {
	infos := make([]*schema.ToolInfo, 0, len(tools))
	for _, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			log.Fatal("get tool info of fail: ", err)
		}
		infos = append(infos, info)
	}

	ncm, err := cm.WithTools(infos)
	if err != nil {
		log.Fatal("bind tools fail: ", err)
	}
	return ncm
}
