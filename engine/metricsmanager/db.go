package metricsmanager

import (
	"context"
	"database/sql"

	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/util"
)

const engineVersion = 1

// DB handles updating metrics
type DB struct {
	db   *sql.DB
	lock *processinglock.Lock

	highAlertID *sql.Stmt
	lowAlertID  *sql.Stmt

	recentlyClosed *sql.Stmt
	scanAlerts     *sql.Stmt
	insertMetrics  *sql.Stmt
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.MetricsManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Version: engineVersion,
		Type:    processinglock.TypeMetrics,
	})
	if err != nil {
		return nil, err
	}

	p := &util.Prepare{Ctx: ctx, DB: db}

	return &DB{
		db:   db,
		lock: lock,

		highAlertID: p.P(`select max(id) from alerts where status = 'closed'`),
		lowAlertID:  p.P(`select min(id) from alerts where status = 'closed'`),

		recentlyClosed: p.P(`
			select distinct log.alert_id
			from alert_logs log
			left join alert_metrics m on m.id = log.alert_id
			where m isnull and log.event = 'closed' and log.timestamp >= now() - '1 hour'::interval
			limit 500
		`),

		scanAlerts: p.P(`
			select a.id
			from alerts a
			left join alert_metrics m on m.id = a.id
			where m isnull and a.status = 'closed' and a.id between $1 and $2
		`),

		insertMetrics: p.P(`
			insert into alert_metrics
			select
				a.id,
				a.service_id,
				(select timestamp - a.created_at from alert_logs where alert_id = a.id and event = 'acknowledged' order by timestamp limit 1),
				(select timestamp - a.created_at from alert_logs where alert_id = a.id and event = 'closed'       order by timestamp limit 1),
				(select count(*) > 1             from alert_logs where alert_id = a.id and event = 'escalated')
			from alerts a
			where a.id = any($1) and a.service_id is not null
		`),
	}, p.Err
}
