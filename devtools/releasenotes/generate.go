package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

func (g *Generator) lastMessage() openai.ChatCompletionMessage {
	if len(g.state) == 0 {
		return openai.ChatCompletionMessage{}
	}

	return g.state[len(g.state)-1]
}

func (g *Generator) Next(ctx context.Context, userInput string) (string, bool, error) {
	g.state = append(g.state, openai.ChatCompletionMessage{
		Content: userInput,
		Role:    "user",
	})

	for {
		err := g.run(ctx)
		if err != nil {
			return "", false, err
		}
		fc := g.lastMessage().FunctionCall
		if fc != nil {
			if fc.Name == g.finalFn {
				return fc.Arguments, true, nil
			}

			result := g.RunFunction(ctx, fc.Name, fc.Arguments)
			g.state = append(g.state, openai.ChatCompletionMessage{
				Content: result,
				Role:    "function",
				Name:    fc.Name,
			})
			continue
		}

		switch g.lastMessage().Role {
		case "assistant":
			return g.lastMessage().Content, false, nil
		default:
			return "", false, fmt.Errorf("unknown or unexpected message role: %s", g.lastMessage().Role)
		}
	}
}

func (g *Generator) run(ctx context.Context) error {
	req := openai.ChatCompletionRequest{
		Model:     g.Model,
		Messages:  g.state,
		Functions: g.fnDefs,
	}

	resp, err := g.AI.CreateChatCompletion(ctx, req)
	if err != nil {
		return err
	}
	if resp.Choices == nil || len(resp.Choices) == 0 {
		return errors.New("no choices returned")
	}
	c := resp.Choices[0]
	switch c.FinishReason {
	case openai.FinishReasonStop, openai.FinishReasonFunctionCall:
		g.state = append(g.state, resp.Choices[0].Message)
		return nil
	case openai.FinishReasonLength:
		g.state = append(g.state, resp.Choices[0].Message)
		return errors.New("max length reached")
	case openai.FinishReasonContentFilter:
		return errors.New("content filter triggered")
	default:
		return fmt.Errorf("unknown finish reason: %s", c.FinishReason)
	}
}
