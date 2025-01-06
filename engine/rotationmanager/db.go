package rotationmanager

import (
	"context"
	"database/sql"

	"github.com/target/goalert/engine/processinglock"
)

// DB manages rotations in Postgres.
type DB struct {
	lock *processinglock.Lock
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.RotationManager" }

// NewDB will create a new DB, preparing all statements necessary.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Type:    processinglock.TypeRotation,
		Version: 2,
	})
	if err != nil {
		return nil, err
	}

	return &DB{
		lock: lock,
	}, nil
}
