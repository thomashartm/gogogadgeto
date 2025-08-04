package main

import (
	"context"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"log"
	"time"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino-ext/components/tool/commandline/sandbox"
	"github.com/cloudwego/eino/components/tool"
)

func newSandbox(ctx context.Context) *sandbox.DockerSandbox {
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
func newDummyTool() tool.BaseTool {
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

func newCommandLineTool(ctx context.Context, sb commandline.Operator) []tool.BaseTool {
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

func bindTools(ctx context.Context, cm model.ToolCallingChatModel, tools []tool.BaseTool) model.ToolCallingChatModel {
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
