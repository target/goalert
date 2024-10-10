package csp

import (
	"context"
)

type nonceval struct{}

// WithNonce will add a nonce value to the context.
func WithNonce(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, nonceval{}, value)
}

// NonceValue will return the nonce value from the context.
func NonceValue(ctx context.Context) string {
	v := ctx.Value(nonceval{})
	if v == nil {
		return ""
	}
	return v.(string)
}
