package imapmanager

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/config"
	"github.com/target/goalert/engine/processinglock"
)

// DB handles IMAP email polling and alert creation.
type DB struct {
	db   *sql.DB
	lock *processinglock.Lock

	alertStore *alert.Store
	cfg        config.Source

	logger *slog.Logger
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.IMAPManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB, alertStore *alert.Store, cfg config.Source, log *slog.Logger) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Version: 1,
		Type:    processinglock.TypeIMAP,
	})
	if err != nil {
		return nil, err
	}

	return &DB{
		db:         db,
		lock:       lock,
		logger:     log,
		alertStore: alertStore,
		cfg:        cfg,
	}, nil
}
