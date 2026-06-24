package statusmgr

import (
	"context"
	"database/sql"

	"github.com/target/goalert/config"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/notification/nfydest"
)

// DB manages outgoing status updates.
type DB struct {
	lock *processinglock.Lock

	reg    *nfydest.Registry
	cfgSrc config.Source
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.StatusUpdateManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB, reg *nfydest.Registry, cfg config.Source) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Type:    processinglock.TypeStatusUpdate,
		Version: 5,
	})
	if err != nil {
		return nil, err
	}

	return &DB{
		lock:   lock,
		reg:    reg,
		cfgSrc: cfg,
	}, nil
}
