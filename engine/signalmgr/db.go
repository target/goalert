package signalmgr

import (
	"context"
	"database/sql"

	"github.com/target/goalert/engine/processinglock"
)

// DB schedules outgoing signal messages.
type DB struct {
	lock *processinglock.Lock
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.SignalManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Type:    processinglock.TypeSignals,
		Version: 1,
	})
	if err != nil {
		return nil, err
	}

	return &DB{
		lock: lock,
	}, nil
}
