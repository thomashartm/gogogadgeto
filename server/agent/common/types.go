package common

import "github.com/cloudwego/eino/schema"

type State struct {
	History   []*schema.Message
	UserInput string
	Name      string
}
