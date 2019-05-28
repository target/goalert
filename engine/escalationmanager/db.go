package escalationmanager

import (
	"context"
	"database/sql"
	alertlog "github.com/target/goalert/alert/log"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/util"
)

// DB handles updating escalation policies.
type DB struct {
	lock *processinglock.Lock

	cleanupNoSteps *sql.Stmt

	lockStmt     *sql.Stmt
	updateOnCall *sql.Stmt

	newPolicies      *sql.Stmt
	deletedSteps     *sql.Stmt
	normalEscalation *sql.Stmt

	log alertlog.Store
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.EscalationManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB, log alertlog.Store) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Version: 3,
		Type:    processinglock.TypeEscalation,
	})
	if err != nil {
		return nil, err
	}

	p := &util.Prepare{Ctx: ctx, DB: db}

	return &DB{
		log:  log,
		lock: lock,

		lockStmt: p.P(`lock escalation_policy_steps in share mode`),

		updateOnCall: p.P(`
			with on_call as (
				select
					step.id step_id,
					coalesce(act.user_id, part.user_id, sched.user_id) user_id
				from escalation_policy_steps step
				join escalation_policy_actions act on act.escalation_policy_step_id = step.id
				left join rotation_state rState on rState.rotation_id = act.rotation_id
				left join rotation_participants part on part.id = rState.rotation_participant_id
				left join schedule_on_call_users sched on sched.schedule_id = act.schedule_id and sched.end_time isnull
				where coalesce(act.user_id, part.user_id, sched.user_id) notnull
			), ended as (
				select
				ep_step_id step_id,
					user_id
				from ep_step_on_call_users
				where end_time isnull
				except
				select step_id, user_id
				from on_call
			), _end as (
				update ep_step_on_call_users ep
				set end_time = now()
				from ended
				where
					ep.ep_step_id = ended.step_id and
					ep.user_id = ended.user_id and
					ep.end_time isnull
			) 
			insert into ep_step_on_call_users (ep_step_id, user_id)
			select step_id, user_id
			from on_call
			on conflict do nothing
			returning ep_step_id, user_id
		`),

		cleanupNoSteps: p.P(`
			delete from escalation_policy_state state
			using escalation_policies pol
			where
				state.escalation_policy_step_id isnull and
				pol.id = state.escalation_policy_id and
				pol.step_count = 0
		`),

		newPolicies: p.P(`
			with to_escalate as (
				select alert_id, step.id ep_step_id, step.delay, step.escalation_policy_id, a.service_id
				from escalation_policy_state state
				join escalation_policy_steps step on
					step.escalation_policy_id = state.escalation_policy_id and
					step.step_number = 0
				join alerts a on a.id = state.alert_id and (a.status = 'triggered' or state.force_escalation)
				where state.last_escalation isnull
				for update skip locked
				limit 1000
			), _cycles as (
				insert into notification_policy_cycles (alert_id, user_id)
				select esc.alert_id, on_call.user_id
				from to_escalate esc
				join ep_step_on_call_users on_call on
					on_call.end_time isnull and
					on_call.ep_step_id = esc.ep_step_id
			), _channels as (
				insert into outgoing_messages (message_type, alert_id, service_id, escalation_policy_id, channel_id)
				select
					cast('alert_notification' as enum_outgoing_messages_type),
					esc.alert_id,
					esc.service_id,
					esc.escalation_policy_id,
					act.channel_id
				from to_escalate esc
				join escalation_policy_actions act on
					act.channel_id notnull and
					act.escalation_policy_step_id = esc.ep_step_id
			)
			update escalation_policy_state state
			set
				last_escalation = now(),
				next_escalation = now() + (cast(esc.delay as text)||' minutes')::interval,
				escalation_policy_step_id = esc.ep_step_id,
				force_escalation = false
			from
				to_escalate esc
			where
				state.alert_id = esc.alert_id
			returning state.alert_id
		`),

		deletedSteps: p.P(`
			with to_escalate as (
				select
					alert_id,
					step.id ep_step_id,
					step.step_number,
					step.delay,
					state.escalation_policy_step_number >= ep.step_count repeated,
					a.service_id,
					step.escalation_policy_id
				from escalation_policy_state state
				join alerts a on a.id = state.alert_id and (a.status = 'triggered' or state.force_escalation)
				join escalation_policies ep on ep.id = state.escalation_policy_id
				join escalation_policy_steps step on
					step.escalation_policy_id = state.escalation_policy_id and
					step.step_number = CASE
						WHEN state.escalation_policy_step_number >= ep.step_count THEN 0
						ELSE state.escalation_policy_step_number
						END
				where
					state.last_escalation notnull and
					escalation_policy_step_id isnull
				for update skip locked
				limit 100
			), _cycles as (
				insert into notification_policy_cycles (alert_id, user_id)
				select esc.alert_id, on_call.user_id
				from to_escalate esc
				join ep_step_on_call_users on_call on
					on_call.end_time isnull and
					on_call.ep_step_id = esc.ep_step_id
			), _channels as (
				insert into outgoing_messages (message_type, alert_id, service_id, escalation_policy_id, channel_id)
				select
					cast('alert_notification' as enum_outgoing_messages_type),
					esc.alert_id,
					esc.service_id,
					esc.escalation_policy_id,
					act.channel_id
				from to_escalate esc
				join escalation_policy_actions act on
					act.channel_id notnull and
					act.escalation_policy_step_id = esc.ep_step_id
			)
			update escalation_policy_state state
			set
				last_escalation = now(),
				next_escalation = now() + (cast(esc.delay as text)||' minutes')::interval,
				escalation_policy_step_number = esc.step_number,
				escalation_policy_step_id = esc.ep_step_id,
				force_escalation = false
			from
				to_escalate esc
			where
				state.alert_id = esc.alert_id
			returning esc.alert_id, esc.repeated, esc.step_number
		`),

		normalEscalation: p.P(`
			with to_escalate as (
				select
					alert_id,
					nextStep.id ep_step_id,
					nextStep.delay,
					nextStep.step_number,
					force_escalation forced,
					oldStep.delay old_delay,
					oldStep.step_number + 1 >= ep.step_count repeated,
					nextStep.escalation_policy_id,
					a.service_id
				from escalation_policy_state state
				join alerts a on a.id = state.alert_id and (a.status = 'triggered' or state.force_escalation)
				join escalation_policies ep on ep.id = state.escalation_policy_id
				join escalation_policy_steps oldStep on oldStep.id = escalation_policy_step_id
				join escalation_policy_steps nextStep on
					nextStep.escalation_policy_id = state.escalation_policy_id and
					nextStep.step_number = CASE
						WHEN oldStep.step_number + 1 < ep.step_count THEN
							oldStep.step_number + 1
						WHEN force_escalation OR ep.repeat = -1 THEN 0
						WHEN state.loop_count < ep.repeat THEN 0
						ELSE -1
					END
				where
					state.last_escalation notnull and
					escalation_policy_step_id notnull and
					(next_escalation < now() or force_escalation)
				order by next_escalation - now()
				for update skip locked
				limit 500
			), _cycles as (
				insert into notification_policy_cycles (alert_id, user_id)
				select esc.alert_id, on_call.user_id
				from to_escalate esc
				join ep_step_on_call_users on_call on
					on_call.end_time isnull and
					on_call.ep_step_id = esc.ep_step_id
			), _channels as (
				insert into outgoing_messages (message_type, alert_id, service_id, escalation_policy_id, channel_id)
				select
					cast('alert_notification' as enum_outgoing_messages_type),
					esc.alert_id,
					esc.service_id,
					esc.escalation_policy_id,
					act.channel_id
				from to_escalate esc
				join escalation_policy_actions act on
					act.channel_id notnull and
					act.escalation_policy_step_id = esc.ep_step_id
			)
			update escalation_policy_state state
			set
				last_escalation = now(),
				next_escalation = now() + (cast(esc.delay as text)||' minutes')::interval,
				escalation_policy_step_number = esc.step_number,
				escalation_policy_step_id = esc.ep_step_id,
				loop_count = CASE WHEN esc.repeated THEN loop_count + 1 ELSE loop_count END,
				force_escalation = false
			from
				to_escalate esc
			where
				state.alert_id = esc.alert_id
			returning esc.alert_id, esc.repeated, esc.step_number, esc.old_delay, esc.forced
		`),
	}, p.Err
}
