package schedulemanager

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/util"
)

// DB will manage schedules and schedule rules in Postgres.
type DB struct {
	lock *processinglock.Lock

	overrides   *sql.Stmt
	rules       *sql.Stmt
	currentTime *sql.Stmt
	getOnCall   *sql.Stmt
	endOnCall   *sql.Stmt
	startOnCall *sql.Stmt
	data        *sql.Stmt
	updateData  *sql.Stmt

	schedTZ *sql.Stmt

	scheduleOnCallNotification *sql.Stmt

	migrateSchedIDs []uuid.UUID
	migrateMap      map[uuid.UUID]uuid.UUID
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.ScheduleManager" }

// NewDB will create a new DB instance, preparing all statements.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Type:    processinglock.TypeSchedule,
		Version: 3,
	})
	if err != nil {
		return nil, err
	}
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &DB{
		lock: lock,

		overrides: p.P(`
			select
				add_user_id,
				remove_user_id,
				tgt_schedule_id
			from user_overrides
			where now() between start_time and end_time
		`),
		data:       p.P(`select schedule_id, data from schedule_data where data notnull for update`),
		updateData: p.P(`update schedule_data set data = $2 where schedule_id = $1`),
		schedTZ:    p.P(`select id, time_zone from schedules`),
		rules: p.P(`
			select
				rule.schedule_id,
				ARRAY[
					sunday,
					monday,
					tuesday,
					wednesday,
					thursday,
					friday,
					saturday
				],
				start_time,
				end_time,
				coalesce(rule.tgt_user_id, part.user_id)
			from schedule_rules rule
			left join rotation_state rState on rState.rotation_id = rule.tgt_rotation_id
			left join rotation_participants part on part.id = rState.rotation_participant_id
			where
				coalesce(rule.tgt_user_id, part.user_id) notnull
		`),
		getOnCall: p.P(`
			select schedule_id, user_id
			from schedule_on_call_users
			where
				end_time isnull
		`),
		startOnCall: p.P(`
			insert into schedule_on_call_users (schedule_id, start_time, user_id)
			select $1, now(), $2 from users where id = $2
		`),
		endOnCall: p.P(`
			update schedule_on_call_users
			set end_time = now()
			where
				schedule_id = $1 and
				user_id = $2 and
				end_time isnull
		`),
		scheduleOnCallNotification: p.P(`
			insert into outgoing_messages (id, message_type, channel_id, schedule_id) values ($1, 'schedule_on_call_notification', $2, $3)
		`),
		currentTime: p.P(`select now()`),
	}, p.Err
}
