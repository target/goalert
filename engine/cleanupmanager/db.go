package cleanupmanager

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/engine/processinglock"
)

// DB handles updating escalation policies.
type DB struct {
	db   *sql.DB
	lock *processinglock.Lock

	alertStore *alert.Store

	logger *slog.Logger
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.CleanupManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB, alertstore *alert.Store, log *slog.Logger) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Version: 1,
		Type:    processinglock.TypeCleanup,
	})
	if err != nil {
		return nil, err
	}

	return &DB{
		db:     db,
		lock:   lock,
		logger: log,

		alertStore: alertstore,
	}, nil
}
