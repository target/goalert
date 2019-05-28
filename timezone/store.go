package timezone

import (
	"context"
	"database/sql"
)

type Store struct {
	db *sql.DB
}

func NewStore(ctx context.Context, db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}
