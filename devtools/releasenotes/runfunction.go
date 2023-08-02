package main

import (
	"context"
	"fmt"
	"log"
)

func (g *Generator) RunFunction(ctx context.Context, name, args string) string {
	log.Println("RunFunction", name, args)
	fn, ok := g.fns[name]
	if !ok {
		return fmt.Sprintf("error: unknown function: %s", name)
	}
	res, err := fn(ctx, []byte(args))
	if err != nil {
		return fmt.Sprintf("error: %s", err)
	}
	return res
}
