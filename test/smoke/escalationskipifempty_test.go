package smoke

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

// escalationSkipIfEmptySQL builds a two-step policy where the first step points
// at a schedule with no one on call -- the weekday-only rotation on a weekend --
// and the second step pages a real user.
//
// skipIfEmpty is applied to the first step by the caller.
const escalationSkipIfEmptySQL = `
insert into users (id, name, email, role)
values
	({{uuid "user"}}, 'bob', 'joe', 'user');

insert into user_contact_methods (id, user_id, name, type, value)
values
	({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});

insert into user_notification_rules (user_id, contact_method_id, delay_minutes)
values
	({{uuid "user"}}, {{uuid "cm1"}}, 0);

-- A schedule with no rules, so no one is ever on call for it.
insert into schedules (id, name, description, time_zone)
values
	({{uuid "sched"}}, 'weekday only', 'nobody on call', 'UTC');

insert into escalation_policies (id, name, repeat)
values
	({{uuid "eid"}}, 'esc policy', 0);

insert into escalation_policy_steps (id, escalation_policy_id, delay, step_number, skip_if_empty)
values
	({{uuid "es1"}}, {{uuid "eid"}}, 60, 0, {{.SkipIfEmpty}}),
	({{uuid "es2"}}, {{uuid "eid"}}, 60, 1, false);

insert into escalation_policy_actions (escalation_policy_step_id, schedule_id)
values
	({{uuid "es1"}}, {{uuid "sched"}});

insert into escalation_policy_actions (escalation_policy_step_id, user_id)
values
	({{uuid "es2"}}, {{uuid "user"}});

insert into services (id, escalation_policy_id, name)
values
	({{uuid "sid"}}, {{uuid "eid"}}, 'service');

insert into alerts (service_id, summary, dedup_key)
values
	({{uuid "sid"}}, 'testing', 'auto:1:foo');
`

// TestEscalationSkipIfEmpty checks that a step with skip_if_empty set escalates
// immediately when it resolves to no one, rather than burning its delay.
func TestEscalationSkipIfEmpty(t *testing.T) {
	t.Parallel()

	sql := strings.ReplaceAll(escalationSkipIfEmptySQL, "{{.SkipIfEmpty}}", "true")

	h := harness.NewHarness(t, sql, "add-ep-step-skip-if-empty")
	defer h.Close()

	// Step 0 has no one on call, so the alert should fall through to step 1 and
	// page the user without waiting out step 0's 60 minute delay.
	h.Twilio(t).Device(h.Phone("1")).ExpectSMS("testing")
}

// TestEscalationSkipIfEmpty_Disabled is the control: with skip_if_empty unset
// (the default), the alert waits out the empty step's delay as it always has.
func TestEscalationSkipIfEmpty_Disabled(t *testing.T) {
	t.Parallel()

	sql := strings.ReplaceAll(escalationSkipIfEmptySQL, "{{.SkipIfEmpty}}", "false")

	h := harness.NewHarness(t, sql, "add-ep-step-skip-if-empty")
	defer h.Close()

	// Nothing should arrive while the alert sits on the empty step.
	h.Twilio(t).WaitAndAssert()

	h.FastForward(61 * time.Minute)
	h.Twilio(t).Device(h.Phone("1")).ExpectSMS("testing")
}

// TestEscalationSkipIfEmpty_SingleStep guards against skipping on a policy with
// nowhere to escalate to. A single-step policy wraps straight back to itself, so
// skipping there would consume a repeat with no elapsed time -- burning the
// alert's only retry instead of leaving it to fire once the delay is up.
func TestEscalationSkipIfEmpty_SingleStep(t *testing.T) {
	t.Parallel()

	const sql = `
insert into users (id, name, email, role)
values
	({{uuid "user"}}, 'bob', 'joe', 'user');

insert into schedules (id, name, description, time_zone)
values
	({{uuid "sched"}}, 'weekday only', 'nobody on call', 'UTC');

insert into escalation_policies (id, name, repeat)
values
	({{uuid "eid"}}, 'esc policy', 1);

insert into escalation_policy_steps (id, escalation_policy_id, delay, step_number, skip_if_empty)
values
	({{uuid "es1"}}, {{uuid "eid"}}, 60, 0, true);

insert into escalation_policy_actions (escalation_policy_step_id, schedule_id)
values
	({{uuid "es1"}}, {{uuid "sched"}});

insert into services (id, escalation_policy_id, name)
values
	({{uuid "sid"}}, {{uuid "eid"}}, 'service');

insert into alerts (service_id, summary, dedup_key)
values
	({{uuid "sid"}}, 'testing', 'auto:1:foo');
`

	h := harness.NewHarness(t, sql, "add-ep-step-skip-if-empty")
	defer h.Close()

	// Let the engine settle; nothing should be sent, since no one is on-call.
	h.Twilio(t).WaitAndAssert()

	var loopCount int
	var secsUntilNext float64
	err := h.App().DB().QueryRowContext(context.Background(), `
		select loop_count, extract(epoch from (next_escalation - now()))
		from escalation_policy_state
	`).Scan(&loopCount, &secsUntilNext)
	require.NoError(t, err)

	// The repeat must still be available, and the alert must be waiting out the
	// full delay rather than having already looped.
	require.Equal(t, 0, loopCount, "repeat consumed with no elapsed time")
	require.Greater(t, secsUntilNext, 60.0, "step delay was skipped with nowhere to escalate to")
}
