package cleanupmanager

import (
	"context"
	"database/sql"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/engine/processinglock"
)

// DB handles updating escalation policies.
type DB struct {
	db   *sql.DB
	lock *processinglock.Lock

	alertStore *alert.Store

	logIndex int64
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.CleanupManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB, alertstore *alert.Store) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Version: 1,
		Type:    processinglock.TypeCleanup,
	})
	if err != nil {
		return nil, err
	}

	return &DB{
		db:   db,
		lock: lock,

		alertStore: alertstore,
	}, nil
}
