package signalsmanager

import (
	"context"
	"database/sql"

	"github.com/target/goalert/engine/processinglock"
)

const engineVersion = 3

// DB handles updating metrics
type DB struct {
	lock *processinglock.Lock
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.SignalsManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Version: engineVersion,
		Type:    processinglock.TypeSignals,
	})
	if err != nil {
		return nil, err
	}

	return &DB{
		lock: lock,
	}, nil
}
