package basic

import (
	"context"

	"github.com/target/goalert/ctxlock"
)

// Provider implements the auth.IdentityProvider interface.
type Provider struct {
	b *Store

	lim *ctxlock.IDLocker[string]
}

// NewProvider creates a new Provider with the associated config.
func NewProvider(ctx context.Context, store *Store) (*Provider, error) {
	return &Provider{
		b:   store,
		lim: ctxlock.NewIDLocker[string](ctxlock.Config{MaxHeld: 1}),
	}, nil
}
