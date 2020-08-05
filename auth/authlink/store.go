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

func (s *Store) Create(ctx context.Context) (*Status, error) {
	return nil, nil
}

func (s *Store) StatusByID(ctx context.Context, id string) (*Status, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
func (s *Store) Verify(ctx context.Context, id, code string) (bool, error) {
	return false, nil
}
