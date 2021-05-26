package statusupdatemanager

import (
	"context"
	"database/sql"

	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/util"
)

// DB manages outgoing status updates.
type DB struct {
	lock *processinglock.Lock

	insertMessages *sql.Stmt

	needsUpdate   *sql.Stmt
	insertMessage *sql.Stmt
	updateStatus  *sql.Stmt
	cleanupClosed *sql.Stmt
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.StatusUpdateManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Type:    processinglock.TypeStatusUpdate,
		Version: 3,
	})
	if err != nil {
		return nil, err
	}
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &DB{
		lock: lock,

		needsUpdate: p.P(`
			select id, channel_id, contact_method_id, alert_id, a.status
			from alert_status_subscriptions sub
			join alerts a on a.id = sub.alert_id and a.status != sub.last_alert_status
			limit 100
			for update skip locked
		`),
		insertMessage: p.P(`insert into outgoing_messages ...`),
		updateStatus:  p.P(`update alert_status_subscriptions set last_alert_status = $2 where id = $1`),
		cleanupClosed: p.P(`delete from alert_status_subscriptions where id = $1`),

		// - get a subset of last_status != current_status
		// - insert messages for each
		// - if new status is closed, delete row
		// - else update to current status
		//
		// - message module, when there are multiple pending messages for the same alert/destination, drop all but the most recent

	}, p.Err
}
