package cleanupmanager

import (
	"context"
	"database/sql"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/util"
)

// DB handles updating escalation policies.
type DB struct {
	db   *sql.DB
	lock *processinglock.Lock

	keys keyring.Keys

	orphanSlackChan *sql.Stmt
	deleteChan      *sql.Stmt
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.CleanupManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB, keys keyring.Keys) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Version: 1,
		Type:    processinglock.TypeCleanup,
	})
	if err != nil {
		return nil, err
	}

	p := &util.Prepare{Ctx: ctx, DB: db}

	return &DB{
		db:   db,
		lock: lock,

		keys: keys,

		orphanSlackChan: p.P(`
			select
				id, meta->>'tok'
			from notification_channels
			where
				type = 'SLACK' and
				id not in (select channel_id from escalation_policy_actions where channel_id notnull)
			order by created_at
			limit 15
			for update skip locked
		`),
		deleteChan: p.P(`delete from notification_channels where id = any($1)`),
	}, p.Err
}
