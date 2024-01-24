package statusmgr

import (
	"context"
	"database/sql"

	"github.com/target/goalert/engine/processinglock"
)

// DB manages outgoing status updates.
type DB struct {
	lock *processinglock.Lock
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.StatusUpdateManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Type:    processinglock.TypeStatusUpdate,
		Version: 5,
	})
	if err != nil {
		return nil, err
	}

	return &DB{
		lock: lock,
	}, nil
}
