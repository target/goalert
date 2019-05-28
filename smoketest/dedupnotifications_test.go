package smoketest

import (
	"github.com/target/goalert/smoketest/harness"
	"testing"
	"time"
)

// TestDedupNotifications tests that if a single contact method is
// used multiple times in a user's notification rules and if engine
// experiences a disruption and resumes after the notification rule delay,
// that only a single notification is generated.
func TestDedupNotifications(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "user"}}, 'bob', 'joe');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "user"}}, {{uuid "cm1"}}, 1),
		({{uuid "user"}}, {{uuid "cm1"}}, 2);

	insert into escalation_policies (id, name) 
	values
		({{uuid "eid"}}, 'esc policy');
	insert into escalation_policy_steps (id, escalation_policy_id) 
	values
		({{uuid "esid"}}, {{uuid "eid"}});
	insert into escalation_policy_actions (escalation_policy_step_id, user_id) 
	values 
		({{uuid "esid"}}, {{uuid "user"}});

	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');

	insert into alerts (service_id, description) 
	values
		({{uuid "sid"}}, 'testing');

`

	h := harness.NewHarness(t, sql, "escalation-policy-step-reorder")
	defer h.Close()

	h.Delay(time.Second * 15)

	//Test that after 3 minutes, only 1 notification is generated
	h.FastForward(time.Minute * 3)

	h.Twilio().Device(h.Phone("1")).ExpectSMS("testing")
}
