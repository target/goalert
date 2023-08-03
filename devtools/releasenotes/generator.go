package main

import (
	"context"
	"encoding/json"

	"github.com/sashabaranov/go-openai"
)

type Generator struct {
	*Config

	state []openai.ChatCompletionMessage

	fns     map[string]func(context.Context, json.RawMessage) (string, error)
	fnDefs  []*openai.FunctionDefine
	finalFn string
}
type Note struct {
	ID   string
	Note string
}

func NewGenerator(cfg *Config, prompt string) *Generator {
	return &Generator{
		Config: cfg,
		fns:    make(map[string]func(context.Context, json.RawMessage) (string, error)),
		state: []openai.ChatCompletionMessage{
			{Role: "system", Content: prompt},
		},
	}
}
