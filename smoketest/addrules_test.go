package smoketest

import (
	"github.com/target/goalert/smoketest/harness"
	"testing"
	"time"
)

/*
# Add Rules Mid-Cycle

This tests for the following behavior when a notification rule is added during a policy.

1. Rules do not retro-actively trigger mid-cycle.
1. New rules take effect, if they occur in the future (mid-cycle).

## Procedure

1. Create 2 rules, 1 immediate, one after 2 minutes.
1. Check the immediate fired.
1. After 2 minutes, check the 2 minute rule fired.
1. Add a rule for 1 minute, and one for 3 minutes.
1. Check that the 1 minute didn't fire.
1. After another minute (3+ total) ensure the 3 minute rule fired.
1. Escalate the alert
1. Ensure the immediate fired.
1. After 1 minute, ensure the 1 minute rule fired.
*/
func TestAddRules(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "uid"}}, 'bob', 'joe');

	insert into user_contact_methods (id, user_id, name, type, value) 
	values
    	({{uuid "cid"}}, {{uuid "uid"}}, 'personal', 'SMS', {{phone "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "uid"}}, {{uuid "cid"}}, 0),
		({{uuid "uid"}}, {{uuid "cid"}}, 60);

	insert into escalation_policies (id, name, repeat) 
	values 
		({{uuid "eid"}}, 'esc policy', -1);
	insert into escalation_policy_steps (id, escalation_policy_id, delay) 
	values 
		({{uuid "esid"}}, {{uuid "eid"}}, 300);

	insert into escalation_policy_actions (escalation_policy_step_id, user_id) 
	values
		({{uuid "esid"}}, {{uuid "uid"}});

	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');

	insert into alerts (service_id, description) 
	values
		({{uuid "sid"}}, 'testing');

`
	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	tw := h.Twilio()
	d := tw.Device(h.Phone("1"))
	d.ExpectSMS("testing")
	tw.WaitAndAssert()

	h.FastForward(30 * time.Minute)

	// ADD RULES
	h.AddNotificationRule(h.UUID("uid"), h.UUID("cid"), 30)
	h.AddNotificationRule(h.UUID("uid"), h.UUID("cid"), 90)

	h.FastForward(30 * time.Minute)

	d.ExpectSMS("testing")
	tw.WaitAndAssert()

	h.FastForward(30 * time.Minute)

	d.ExpectSMS("testing")
	tw.WaitAndAssert()

	h.Escalate(1, 0)

	d.ExpectSMS("testing")
	tw.WaitAndAssert()

	h.FastForward(30 * time.Minute)

	d.ExpectSMS("testing")
}
