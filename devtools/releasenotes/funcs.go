package main

import (
	"context"
	"encoding/json"

	"github.com/sashabaranov/go-openai"
)

func (g *Generator) AddFunc(def *openai.FunctionDefine, fn func(context.Context, json.RawMessage) (string, error)) {
	g.fns[def.Name] = fn
	g.fnDefs = append(g.fnDefs, def)
}
func (g *Generator) SetFinal(def *openai.FunctionDefine) {
	g.fnDefs = append(g.fnDefs, def)
	g.finalFn = def.Name
}
