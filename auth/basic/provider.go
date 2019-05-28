package basic

import (
	"context"
	"database/sql"
)

// Provider implements the auth.IdentityProvider interface.
type Provider struct {
	b *Store
}

// NewProvider creates a new Provider with the associated config.
func NewProvider(ctx context.Context, db *sql.DB) (*Provider, error) {
	b, err := NewStore(ctx, db)
	if err != nil {
		return nil, err
	}

	return &Provider{b: b}, nil
}
