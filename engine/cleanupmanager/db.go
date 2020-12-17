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

	cleanupAlerts   *sql.Stmt
	cleanupAPIKeys  *sql.Stmt
	setTimeout      *sql.Stmt
	cleanupSessions *sql.Stmt

	cleanupAlertLogs *sql.Stmt

	logIndex int
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
		setTimeout:      p.P(`SET LOCAL statement_timeout = 3000`),
		cleanupAlerts:   p.P(`delete from alerts where id = any(select id from alerts where status = 'closed' AND created_at < (now() - $1::interval) order by id limit 100 for update skip locked)`),
		cleanupAPIKeys:  p.P(`update user_calendar_subscriptions set disabled = true where id = any(select id from user_calendar_subscriptions where greatest(last_access, last_update) < (now() - $1::interval) order by id limit 100 for update skip locked)`),
		cleanupSessions: p.P(`DELETE FROM auth_user_sessions WHERE id = any(select id from auth_user_sessions where last_access_at < (now() - '30 days'::interval) LIMIT 100 for update skip locked)`),

		cleanupAlertLogs: p.P(`
			with
				scope as (select id from alert_logs where id > $1 order by id limit 1000),
				id_range as (select min(id), max(id) from scope),
				_delete as (
					delete from alert_logs where id = any(
						select id from alert_logs
						where
							id between (select min from id_range) and (select max from id_range) and
							not exists (select 1 from alerts where alert_id = id)
						for update skip locked
					)
				)
			select id from scope offset 999
		`),
	}, p.Err
}
