package cleanupmanager

import (
	"context"
	"database/sql"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/util"
)

// DB handles updating escalation policies.
type DB struct {
	db   *sql.DB
	lock *processinglock.Lock

	now *sql.Stmt

	userIDs        *sql.Stmt
	cleanupAPIKeys *sql.Stmt
	setTimeout     *sql.Stmt

	schedData    *sql.Stmt
	setSchedData *sql.Stmt

	cleanupSessions *sql.Stmt

	cleanupAlertLogs *sql.Stmt

	alertStore *alert.Store

	logIndex int
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

	p := &util.Prepare{Ctx: ctx, DB: db}

	return &DB{
		db:   db,
		lock: lock,

		now:     p.P(`select now()`),
		userIDs: p.P(`select id from users`),

		// Abort any cleanup operation that takes longer than 3 seconds
		// error will be logged.
		setTimeout:     p.P(`SET LOCAL statement_timeout = 3000`),
		cleanupAPIKeys: p.P(`update user_calendar_subscriptions set disabled = true where id = any(select id from user_calendar_subscriptions where greatest(last_access, last_update) < (now() - $1::interval) order by id limit 100 for update skip locked)`),

		schedData: p.P(`
			select schedule_id, data from schedule_data
			where data notnull and (last_cleanup_at isnull or last_cleanup_at <= now() - '1 month'::interval)
			order by last_cleanup_at asc nulls first
			for update skip locked
			limit 100
		`),
		setSchedData:    p.P(`update schedule_data set last_cleanup_at = now(), data = $2 where schedule_id = $1`),
		cleanupSessions: p.P(`DELETE FROM auth_user_sessions WHERE id = any(select id from auth_user_sessions where last_access_at < (now() - '30 days'::interval) LIMIT 100 for update skip locked)`),

		cleanupAlertLogs: p.P(`
			with
				scope as (select id from alert_logs where id > $1 order by id limit 100),
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
			select id from scope offset 99
		`),

		alertStore: alertstore,
	}, p.Err
}
