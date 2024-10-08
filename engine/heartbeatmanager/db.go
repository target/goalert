package heartbeatmanager

import (
	"context"
	"database/sql"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/util"
)

// DB processes heartbeats.
type DB struct {
	lock *processinglock.Lock

	alertStore *alert.Store

	fetchFailed  *sql.Stmt
	fetchHealthy *sql.Stmt
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.HeartbeatManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB, a *alert.Store) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Type:    processinglock.TypeHeartbeat,
		Version: 2,
	})
	if err != nil {
		return nil, err
	}

	p := &util.Prepare{Ctx: ctx, DB: db}

	return &DB{
		lock:       lock,
		alertStore: a,

		// if checked is still false after processing, we can delete it
		fetchFailed: p.P(`
			with rows as (
				select id
				from heartbeat_monitors
				where
					last_state != 'unhealthy' and
					now() - last_heartbeat >= heartbeat_interval
				limit 250
				for update skip locked
			)
			update heartbeat_monitors mon
			set last_state = 'unhealthy'
			from rows
			where mon.id = rows.id
			returning mon.id, name, service_id, last_heartbeat, coalesce(additional_details, ''), coalesce(disable_reason, '')
		`),
		fetchHealthy: p.P(`
			with rows as (
				select id
				from heartbeat_monitors
				where
					last_state != 'healthy' and
					now() - last_heartbeat < heartbeat_interval
				limit 250
				for update skip locked
			)
			update heartbeat_monitors mon
			set last_state = 'healthy'
			from rows
			where mon.id = rows.id
			returning mon.id, service_id
		`),
	}, p.Err
}
