package npcyclemanager

import (
	"context"
	"database/sql"

	alertlog "github.com/target/goalert/alert/log"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/util"
)

// DB manages user notification cycles in Postgres.
//
// It handles queueing of notifications.
type DB struct {
	db   *sql.DB
	lock *processinglock.Lock

	queueMessages *sql.Stmt
	log           alertlog.Store
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.NotificationCycleManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB, log alertlog.Store) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Type:    processinglock.TypeNPCycle,
		Version: 2,
	})
	if err != nil {
		return nil, err
	}
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &DB{
		log:  log,
		lock: lock,

		// add messages for notification rules who's delay is between the last tick and now.
		//
		// Example:
		// - policy started at 1:00
		// - notifications were sent for 0-minute at 1:00:15 (last tick = 1:00:15)
		// - at 1:01:15 only notification rules with delays between 15 and 75 seconds would be processed/sent
		// Note: since delays are in minutes, the above example would just send the 1 minute rules (60 seconds)
		queueMessages: p.P(`
			with lock_cycles as (
				select
					id,
					alert_id,
					user_id,
					started_at,
					last_tick
				from notification_policy_cycles
				where
					last_tick isnull or
					last_tick < now() - '1 minute'::interval
				order by
					last_tick nulls first,
					started_at
				for update skip locked
				limit 1250
			), deleted as (
				delete from notification_policy_cycles cycle
				using alerts a, lock_cycles lock
				where
					a.status != 'triggered' and a.id = cycle.alert_id and
					cycle.id = lock.id
				returning cycle.id
			), process_cycles as (
				select *
				from lock_cycles lock
				where not exists (
					select null
					from deleted del
					where lock.id = del.id
				)
			), inserted as (
				insert into outgoing_messages (
					message_type,
					contact_method_id,
					alert_id,
					cycle_id,
					user_id,
					service_id,
					escalation_policy_id
				)
				select distinct
					cast('alert_notification' as enum_outgoing_messages_type),
					rule.contact_method_id,
					cycle.alert_id,
					cycle.id,
					rule.user_id,
					a.service_id,
					svc.escalation_policy_id
				from process_cycles cycle
				join alerts a on a.id = cycle.alert_id
				join services svc on svc.id = a.service_id
				join user_notification_rules rule on
					rule.user_id = cycle.user_id and
					(
						cycle.last_tick isnull or
						concat(rule.delay_minutes,' minutes')::interval > (cycle.last_tick - cycle.started_at)
					) and
					concat(rule.delay_minutes,' minutes')::interval <= (now() - cycle.started_at)
				returning cycle_id
			), no_first_notif_sent as (
				select user_id, alert_id
				from process_cycles
				where last_tick isnull and id not in (select cycle_id from inserted)
			), update as (
				update notification_policy_cycles
				set last_tick = greatest(last_tick, now())
				where id in (select id from process_cycles)
			)
			select user_id, alert_id from no_first_notif_sent
		`),
	}, p.Err
}
