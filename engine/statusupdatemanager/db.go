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
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.StatusUpdateManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Type:    processinglock.TypeStatusUpdate,
		Version: 1,
	})
	if err != nil {
		return nil, err
	}
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &DB{
		lock: lock,

		insertMessages: p.P(`
			with rows as (
				select
					log.id,
					log.alert_id,
					usr.alert_status_log_contact_method_id,
					last.user_id,
					log.event = 'closed' is_closed,
					coalesce(last.user_id = log.sub_user_id, false) is_same_user
				from user_last_alert_log last
				join users usr on
					usr.id = last.user_id
				join alert_logs log ON
					last.alert_id = log.alert_id AND
					log.id BETWEEN last.log_id+1 AND last.next_log_id AND
					log.event IN ('acknowledged', 'closed')
				where last.log_id != last.next_log_id
				limit 100
				for update skip locked
			), inserted as (
				insert into outgoing_messages (
					message_type,
					alert_log_id,
					alert_id,
					contact_method_id,
					user_id
				)
				select
					'alert_status_update',
					id,
					alert_id,
					alert_status_log_contact_method_id,
					user_id
				from rows
				where
					alert_status_log_contact_method_id notnull and
					not is_same_user
			), any_closed as (
				select
					bool_or(is_closed) is_closed, user_id, alert_id
				from rows
				group by user_id, alert_id
			), updated as (
				update user_last_alert_log log
				set log_id = next_log_id
				from any_closed c
				where
					not c.is_closed and
					log.user_id = c.user_id and
					log.alert_id = c.alert_id
			)
			delete from user_last_alert_log log
			using any_closed c
			where
				c.is_closed and
				log.user_id = c.user_id and
				log.alert_id = c.alert_id
		`),
	}, p.Err
}
