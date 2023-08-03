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
	gen := NewGenerator(c, prPrompt)

	gen.SetFinal(&openai.FunctionDefine{
		Name:        "output",
		Description: "Output the summary of the current PR.",
		Parameters: &openai.FunctionParams{
			Type: openai.JSONSchemaTypeObject,
			Properties: map[string]*openai.JSONSchemaDefine{
				"type":     {Type: openai.JSONSchemaTypeString, Description: "The type of change made in this PR (e.g., feature, enhancement, security, bugfix, development)."},
				"feature":  {Type: openai.JSONSchemaTypeString, Description: "The base feature that this PR affects."},
				"section":  {Type: openai.JSONSchemaTypeString, Description: "The section of the release notes that this PR should be included in."},
				"summary":  {Type: openai.JSONSchemaTypeString, Description: "A short single-sentence summary of the change."},
				"internal": {Type: openai.JSONSchemaTypeBoolean, Description: "Skip this PR in the release notes. This is useful for changes that are not user-facing, such as documentation changes, dependency updates, dev changes, or refactoring (e.g., React)."},
				"why":      {Type: openai.JSONSchemaTypeString, Description: "A terse reason (few words) this PR is important to end users."},
			},
			Required: []string{"summary", "section", "type", "feature", "internal", "why"},
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
