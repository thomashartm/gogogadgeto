package model

import (
	"context"
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

func CreateAgent() compose.Runnable[string, string] {
	util.LogMessage("=== AGENT CREATION START ===")
	ctx := context.Background()

	agent := composeAgent(ctx, cm, toolsList)

	util.LogMessage("=== AGENT CREATION COMPLETE ===")
	return agent
}

func composeAgent(ctx context.Context,
	cm model.BaseChatModel,
	tools []tool.BaseTool,
) compose.Runnable[string, string] {

}
