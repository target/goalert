package authlink

import (
	"context"
	"database/sql"
	"time"

	"github.com/target/goalert/permission"
)

type Store struct{}

type Status struct {
	ID string

	ClaimCode  string
	VerifyCode string

	CreatedAt  time.Time
	ExpiresAt  time.Time
	ClaimedAt  time.Time
	VerifiedAt time.Time
	AuthedAt   time.Time
}

func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	return &Store{}, nil
}

// Create will create a new auth link for the session associated with ctx.
func (s *Store) Create(ctx context.Context) (*Status, error) {
	return nil, nil
}

// StatusByID will return the current auth link status for the given ID.
func (s *Store) StatusByID(ctx context.Context, id string) (*Status, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Verify will perform verification on the provided auth link with the given code.
func (s *Store) Verify(ctx context.Context, id, code string) (bool, error) {
	return false, nil
}

// Reset will reset any outstanding auth links for the user associated with ctx.
func (s *Store) Reset(ctx context.Context) error {
	return nil
}
