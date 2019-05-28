
-- +migrate Up

LOCK escalation_policy_state, escalation_policy_steps, notification_policy_cycles;

UPDATE escalation_policy_state state
SET next_escalation = last_escalation + cast(cast(step.delay as text)||' minutes' as interval)
FROM escalation_policy_steps step
WHERE next_escalation isnull and step.id = state.escalation_policy_step_id;


-- pre-populate on call
with on_call as (
    select distinct
        step.id step_id,
        coalesce(act.user_id, part.user_id, sched.user_id) user_id
    from escalation_policy_steps step
    join escalation_policy_actions act on act.escalation_policy_step_id = step.id
    left join rotation_state rState on rState.rotation_id = act.rotation_id
    left join rotation_participants part on part.id = rState.rotation_participant_id
    left join schedule_on_call_users sched on sched.schedule_id = act.schedule_id and sched.end_time isnull
    where coalesce(act.user_id, part.user_id, sched.user_id) notnull
), new_on_call as (
    insert into ep_step_on_call_users (ep_step_id, user_id)
    select step_id, user_id
    from on_call
    on conflict do nothing
    returning ep_step_id, user_id
)
insert into notification_policy_cycles (alert_id, user_id)
select state.alert_id, oncall.user_id
from escalation_policy_state state
join alerts a on a.id = state.alert_id and a.status = 'triggered'
join new_on_call oncall on oncall.ep_step_id = state.escalation_policy_step_id
except
select alert_id, user_id
from notification_policy_cycles;

-- +migrate Down

UPDATE escalation_policy_state
SET next_escalation = null;
