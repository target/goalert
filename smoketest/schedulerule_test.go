package smoketest

import (
	"github.com/target/goalert/smoketest/harness"
	"testing"
	"time"
)

// TestScheduleRule performs the following checks:
// - A schedule rule "shift" can end with a past shift in the DB (bug found in dev)
// - Schedule rule time constraints are evaluated correctly
func TestScheduleRule(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "u1"}}, 'bob', 'joe'),
		({{uuid "u2"}}, 'ben', 'josh');

	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "cm1"}}, {{uuid "u1"}}, 'personal', 'SMS', {{phone "1"}}),
		({{uuid "cm2"}}, {{uuid "u2"}}, 'personal', 'SMS', {{phone "2"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "u1"}}, {{uuid "cm1"}}, 0),
		({{uuid "u2"}}, {{uuid "cm2"}}, 0);

	insert into escalation_policies (id, name) 
	values
		({{uuid "eid"}}, 'esc policy');

	insert into escalation_policy_steps (id, escalation_policy_id) 
	values
		({{uuid "esid"}}, {{uuid "eid"}});

	insert into schedules (id, name, description, time_zone)
	values
		({{uuid "sched"}}, 'test', 'test', 'America/Chicago');
	
	insert into schedule_rules (schedule_id, start_time, end_time, tgt_user_id)
	values
		({{uuid "sched"}}, cast((now()-'5 minutes'::interval) at time zone 'America/Chicago' as time without time zone), cast((now()+'5 minutes'::interval) at time zone 'America/Chicago' as time without time zone), {{uuid "u1"}}),
		({{uuid "sched"}}, cast((now()+'5 minutes'::interval) at time zone 'America/Chicago' as time without time zone), cast((now()+'15 minutes'::interval) at time zone 'America/Chicago' as time without time zone), {{uuid "u2"}});

	insert into escalation_policy_actions (escalation_policy_step_id, schedule_id) 
	values 
		({{uuid "esid"}}, {{uuid "sched"}});

	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');
	
	insert into schedule_on_call_users (schedule_id, start_time, end_time, user_id)
	values
		({{uuid "sched"}}, now()-'2 hours'::interval, now()-'1 hour'::interval, {{uuid "u1"}});

`
	h := harness.NewHarness(t, sql, "npcycle-indexes")
	defer h.Close()

	sid := h.UUID("sid")
	u1 := h.UUID("u1")
	u2 := h.UUID("u2")

	h.WaitAndAssertOnCallUsers(sid, u1)

	h.FastForward(10 * time.Minute)

	h.WaitAndAssertOnCallUsers(sid, u2)
}
