package smoketest

import (
	"testing"
	"time"

	"github.com/target/goalert/smoketest/harness"
)

// TestPostCycleRules checks that new rules added after the last
// rule of a policy executes are handled the same way as during a policy cycle.
func TestPostCycleRules(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "uid"}}, 'bob', 'joe');

	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "cid"}}, {{uuid "uid"}}, 'personal', 'SMS', {{phone "1"}}),
		({{uuid "cid2"}}, {{uuid "uid"}}, 'personal2', 'SMS', {{phone "2"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "uid"}}, {{uuid "cid2"}}, 0);

	insert into escalation_policies (id, name, repeat) 
	values 
		({{uuid "eid"}}, 'esc policy', -1);
	insert into escalation_policy_steps (id, escalation_policy_id, delay) 
	values 
		({{uuid "esid"}}, {{uuid "eid"}}, 60);

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

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))
	d2 := tw.Device(h.Phone("2"))

	d2.ExpectSMS("testing")
	tw.WaitAndAssert()

	// ADD RULES
	h.AddNotificationRule(h.UUID("uid"), h.UUID("cid"), 0)
	h.AddNotificationRule(h.UUID("uid"), h.UUID("cid"), 30)

	// ensure no notification for instant rule
	tw.WaitAndAssert()

	h.FastForward(30 * time.Minute)

	d1.ExpectSMS("testing")
	tw.WaitAndAssert()
}
