package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/sashabaranov/go-openai"
)

var inR = bufio.NewReader(os.Stdin)

func (c *Config) SummarizeIssue(ctx context.Context, number int) (string, error) {
	gen := NewGenerator(c, issuePrompt)

	gen.SetFinal(&openai.FunctionDefine{
		Name:        "output_summary",
		Description: "Output the summary of the current issue.",
		Parameters: &openai.FunctionParams{
			Type: openai.JSONSchemaTypeObject,
			Properties: map[string]*openai.JSONSchemaDefine{
				"type":      {Type: openai.JSONSchemaTypeString, Description: "The type of issue (e.g., feature, enhancement, security, bugfix)."},
				"component": {Type: openai.JSONSchemaTypeString, Description: "The component or base feature that this issue affects."},
				"summary":   {Type: openai.JSONSchemaTypeString, Description: "A short summary of the issue."},
			},
			Required: []string{"summary", "type"},
		},
	})

	gen.AddFunc(&openai.FunctionDefine{
		Name:        "issue_summary",
		Description: "Get a summary of a related issue.",
		Parameters: &openai.FunctionParams{
			Type: openai.JSONSchemaTypeObject,
			Properties: map[string]*openai.JSONSchemaDefine{
				"id": {Type: openai.JSONSchemaTypeNumber, Description: "id of the issue to read"},
			},
			Required: []string{"id"},
		},
	}, func(ctx context.Context, data json.RawMessage) (string, error) {
		var args struct {
			ID int
		}
		err := json.Unmarshal(data, &args)
		if err != nil {
			return "", err
		}
		return c.SummarizeIssue(ctx, args.ID)
	})

	info, err := gen.ReadIssue(ctx, number)
	if err != nil {
		panic(err)
	}
	data, err := json.Marshal(info)
	if err != nil {
		panic(err)
	}
	resp, isDone, err := gen.Next(ctx, string(data))
	if err != nil {
		panic(err)
	}
	if !isDone {
		log.Fatalln("expected done, got:", resp)
	}

	return resp, nil
}

func MkUserQuestion(r *bufio.Reader) (*openai.FunctionDefine, func(context.Context, json.RawMessage) (string, error)) {
	return &openai.FunctionDefine{
			Name:        "user_question",
			Description: "Ask a developer to provide more information.",
			Parameters: &openai.FunctionParams{
				Type: openai.JSONSchemaTypeObject,
				Properties: map[string]*openai.JSONSchemaDefine{
					"question": {Type: openai.JSONSchemaTypeString, Description: "The question to ask the user."},
				},
				Required: []string{"question"},
			},
		}, func(ctx context.Context, data json.RawMessage) (string, error) {
			var args struct {
				Question string
			}
			err := json.Unmarshal(data, &args)
			if err != nil {
				return "", err
			}

			fmt.Println(args.Question)
			resp, err := r.ReadString('\n')
			if err != nil {
				return "", err
			}

			return resp, nil
		}
}
