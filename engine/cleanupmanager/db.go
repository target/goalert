package cleanupmanager

import (
	"context"
	"database/sql"

	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/util"
)

// DB handles updating escalation policies.
type DB struct {
	db   *sql.DB
	lock *processinglock.Lock

	cleanupAlerts  *sql.Stmt
	cleanupAPIKeys *sql.Stmt
	setTimeout     *sql.Stmt
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.CleanupManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
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

		// Abort any cleanup operation that takes longer than 3 seconds
		// error will be logged.
		setTimeout:     p.P(`SET LOCAL statement_timeout = 3000`),
		cleanupAlerts:  p.P(`delete from alerts where id = any(select id from alerts where status = 'closed' AND created_at < (now() - $1::interval) order by id limit 100 for update)`),
		cleanupAPIKeys: p.P(`update user_calendar_subscriptions set disabled = true where id = any(select id from user_calendar_subscriptions where last_update < (now() - $1::interval))`),
	}, p.Err
}
