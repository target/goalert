package apikey

import "context"

type contextKey int

const (
	contextKeyPolicy contextKey = iota
)

// PolicyFromContext returns the Policy associated with the given context.
func PolicyFromContext(ctx context.Context) *Policy {
	p, _ := ctx.Value(contextKeyPolicy).(*Policy)
	return p
}

// ContextWithPolicy returns a new context with the given Policy attached.
func ContextWithPolicy(ctx context.Context, p *Policy) context.Context {
	return context.WithValue(ctx, contextKeyPolicy, p)
}
