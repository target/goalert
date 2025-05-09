package smoke

import (
	"fmt"
	"testing"
	"time"

	"github.com/target/goalert/test/smoke/harness"
	"github.com/target/goalert/util/timeutil"
)

// TestScheduleOverrideRemove validates that a "remove" style override correctly removes an existing shift.
//
// - User A is always on-call
// - User B is on-call for a specific time (via rule)
// - User B has an override that removes them from the schedule for a specific time.
func TestScheduleOverrideRemove(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "u1"}}, 'bob', 'joe'),
		({{uuid "u2"}}, 'ben', 'josh');

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
		({{uuid "sched"}}, '00:00', '00:00', {{uuid "u1"}}),
		({{uuid "sched"}}, '09:00', '21:00', {{uuid "u2"}});

	insert into escalation_policy_actions (escalation_policy_step_id, schedule_id) 
	values 
		({{uuid "esid"}}, {{uuid "sched"}});

	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');

`
	h := harness.NewHarness(t, sql, "npcycle-indexes")
	defer h.Close()

	svcID := h.UUID("sid")
	u1 := h.UUID("u1")
	u2 := h.UUID("u2")

	now := h.Now()

	start := now.AddDate(0, 0, 2)
	end := now.AddDate(0, 0, 6)

	h.GraphQLQuery2(fmt.Sprintf(`
	mutation{
	createUserOverride(input:{
		removeUserID: "%s",
		scheduleID:"%s",
		start:"%s",
		end:"%s"}) { id }
	}`,
		u2, h.UUID("sched"), start.Format(time.RFC3339), end.Format(time.RFC3339)))

	h.FastForwardToTime(timeutil.NewClock(0, 0), "America/Chicago")
	h.WaitAndAssertOnCallUsers(svcID, u1)
	h.FastForwardToTime(timeutil.NewClock(9, 0), "America/Chicago")
	h.WaitAndAssertOnCallUsers(svcID, u1, u2)
	h.FastForwardToTime(timeutil.NewClock(21, 0), "America/Chicago")
	h.WaitAndAssertOnCallUsers(svcID, u1)

	h.FastForward(24 * time.Hour * 3)

	h.FastForwardToTime(timeutil.NewClock(0, 0), "America/Chicago")
	h.WaitAndAssertOnCallUsers(svcID, u1)
	h.FastForwardToTime(timeutil.NewClock(9, 0), "America/Chicago")
	h.WaitAndAssertOnCallUsers(svcID, u1)
	h.FastForwardToTime(timeutil.NewClock(21, 0), "America/Chicago")
	h.WaitAndAssertOnCallUsers(svcID, u1)

	h.FastForward(24 * time.Hour * 3)
	h.FastForwardToTime(timeutil.NewClock(0, 0), "America/Chicago")
	h.WaitAndAssertOnCallUsers(svcID, u1)
	h.FastForwardToTime(timeutil.NewClock(9, 0), "America/Chicago")
	h.WaitAndAssertOnCallUsers(svcID, u1, u2)
	h.FastForwardToTime(timeutil.NewClock(21, 0), "America/Chicago")
	h.WaitAndAssertOnCallUsers(svcID, u1)
}
