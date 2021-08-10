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

	latestLogEntry *sql.Stmt
	needsUpdate    *sql.Stmt
	insertMessage  *sql.Stmt
	updateStatus   *sql.Stmt
	deleteSub      *sql.Stmt
	cmWantsUpdates *sql.Stmt
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

		cmWantsUpdates: p.P(`
			select coalesce(u.alert_status_log_contact_method_id = $1, false), u.id
			from user_contact_methods cm
			join users u on u.id = cm.user_id
			where cm.id = $1 and not cm.disabled
		`),

		needsUpdate: p.P(`
			select sub.id, channel_id, contact_method_id, alert_id, (select status from alerts a where a.id = sub.alert_id)
			from alert_status_subscriptions sub
			where sub.last_alert_status != (select status from alerts a where a.id = sub.alert_id)
			limit 1
			for update skip locked
		`),

		insertMessage: p.P(`
			insert into outgoing_messages(
				id,
				message_type,
				channel_id,
				contact_method_id,
				user_id,
				alert_id,
				alert_log_id
			) values ($1, 'alert_status_update', $2, $3, $4, $5, $6)
		`),

		latestLogEntry: p.P(`
			select id, sub_user_id from alert_logs
			where alert_id = $1 and event = $2 and timestamp > now() - '1 hour'::interval
			order by id desc
			limit 1
		`),

		updateStatus: p.P(`update alert_status_subscriptions set last_alert_status = $2 where id = $1`),
		deleteSub:    p.P(`delete from alert_status_subscriptions where id = $1`),
	}, p.Err
}
