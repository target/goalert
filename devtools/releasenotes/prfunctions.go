package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"log"

	"github.com/sashabaranov/go-openai"
)

//go:embed prsummary.txt
var prPrompt string

func (c *Config) SummarizePR(ctx context.Context, number int) (string, error) {
	gen := NewGenerator(c.GH, c.AI, prPrompt)

	gen.SetFinal(&openai.FunctionDefine{
		Name:        "output_summary",
		Description: "Output a summary of the current PR.",
		Parameters: &openai.FunctionParams{
			Type: openai.JSONSchemaTypeObject,
			Properties: map[string]*openai.JSONSchemaDefine{
				"is_dev_only":    {Type: openai.JSONSchemaTypeBoolean, Description: "The PR/change only affects the developers of GoAlert."},
				"is_admin":       {Type: openai.JSONSchemaTypeBoolean, Description: "The PR/change affects admins of GoAlert (e.g., new command line flags, admin-only features, etc...)."},
				"is_user_facing": {Type: openai.JSONSchemaTypeBoolean, Description: "The PR/change affects or adds to the available features/end-user-facing components of GoAlert."},
				"summary":        {Type: openai.JSONSchemaTypeString, Description: "A short summary of the PR."},
			},
			Required: []string{"summary", "is_dev_only", "is_admin", "is_user_facing"},
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

	gen.AddFunc(MkUserQuestion(inR))

	info, err := gen.ReadPR(ctx, number)
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
