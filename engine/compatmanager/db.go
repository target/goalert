package compatmanager

import (
	"context"
	"database/sql"

	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/notification/slack"
)

// DB handles keeping compatibility-related data in sync.
type DB struct {
	db   *sql.DB
	lock *processinglock.Lock

	cs *slack.ChannelSender
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.CompatManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB, cs *slack.ChannelSender) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Version: 1,
		Type:    processinglock.TypeCompat,
	})
	if err != nil {
		return nil, err
	}

	return &DB{
		db:   db,
		lock: lock,
		cs:   cs,
	}, nil
}
