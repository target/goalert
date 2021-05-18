package basic

import (
	"context"
)

// Provider implements the auth.IdentityProvider interface.
type Provider struct {
	b *Store
}

// NewProvider creates a new Provider with the associated config.
func NewProvider(ctx context.Context, store *Store) (*Provider, error) {
	return &Provider{b: store}, nil
}
