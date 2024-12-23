package cleanupmanager

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/util"
)

// DB handles updating escalation policies.
type DB struct {
	db   *sql.DB
	lock *processinglock.Lock

	cleanupAPIKeys *sql.Stmt
	setTimeout     *sql.Stmt

	cleanupSessions *sql.Stmt

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

	p := &util.Prepare{Ctx: ctx, DB: db}

	return &DB{
		db:     db,
		lock:   lock,
		logger: log,

		// Abort any cleanup operation that takes longer than 3 seconds
		// error will be logged.
		setTimeout:     p.P(`SET LOCAL statement_timeout = 3000`),
		cleanupAPIKeys: p.P(`update user_calendar_subscriptions set disabled = true where id = any(select id from user_calendar_subscriptions where greatest(last_access, last_update) < (now() - $1::interval) order by id limit 100 for update skip locked)`),

		cleanupSessions: p.P(`DELETE FROM auth_user_sessions WHERE id = any(select id from auth_user_sessions where last_access_at < (now() - '30 days'::interval) LIMIT 100 for update skip locked)`),

		alertStore: alertstore,
	}, p.Err
}
