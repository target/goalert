package config

import "context"

type contextKey string

var contextKeyConfig = contextKey("config")

// Context returns a new Context that carries the provided Config.
func (cfg Config) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyConfig, cfg)
}

// FromContext will return the Config carried in the provided Context.
//
// It panics if config is not available on the current context.
func FromContext(ctx context.Context) Config {
	return ctx.Value(contextKeyConfig).(Config)
}
