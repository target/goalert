package statusmgr

import (
	"context"
	"database/sql"

	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/notification/nfydest"
)

// DB manages outgoing status updates.
type DB struct {
	lock *processinglock.Lock

	omit []int64
	reg  *nfydest.Registry
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.StatusUpdateManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB, reg *nfydest.Registry) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Type:    processinglock.TypeStatusUpdate,
		Version: 5,
	})
	if err != nil {
		return nil, err
	}

	return &DB{
		lock: lock,
		reg:  reg,
		omit: make([]int64, 0, 100), // pre-allocate for 100, needs to not be nil or the query will fail
	}, nil
}
