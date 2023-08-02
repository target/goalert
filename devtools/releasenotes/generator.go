package main

import (
	"context"
	"encoding/json"

	"github.com/google/go-github/v51/github"
	"github.com/sashabaranov/go-openai"
)

type Generator struct {
	gh *github.Client
	ai *openai.Client

	state []openai.ChatCompletionMessage

	fns     map[string]func(context.Context, json.RawMessage) (string, error)
	fnDefs  []*openai.FunctionDefine
	finalFn string
}
type Note struct {
	ID   string
	Note string
}

func NewGenerator(gh *github.Client, ai *openai.Client, prompt string) *Generator {
	return &Generator{
		gh:  gh,
		ai:  ai,
		fns: make(map[string]func(context.Context, json.RawMessage) (string, error)),
		state: []openai.ChatCompletionMessage{
			{Role: "system", Content: prompt},
		},
	}
}
