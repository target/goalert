package smoketest

import (
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

// TestMultiUser checks that if multiple users are assigned to a policy step,
// they all get their notifications.
func TestMultiUser(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email) 
	values
		({{uuid "u1"}}, 'bob', 'joe'),
		({{uuid "u2"}}, 'ben', 'josh'),
		({{uuid "u3"}}, 'beth', 'jake');

	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "c1"}}, {{uuid "u1"}}, 'personal', 'SMS', {{phone "1"}}),
		({{uuid "c2"}}, {{uuid "u2"}}, 'personal', 'SMS', {{phone "2"}}),
		({{uuid "c3"}}, {{uuid "u3"}}, 'personal', 'SMS', {{phone "3"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "u1"}}, {{uuid "c1"}}, 0),
		({{uuid "u2"}}, {{uuid "c2"}}, 0),
		({{uuid "u3"}}, {{uuid "c3"}}, 0);

	insert into escalation_policies (id, name, repeat) 
	values 
		({{uuid "eid"}}, 'esc policy', -1);
	insert into escalation_policy_steps (id, escalation_policy_id, delay) 
	values 
		({{uuid "esid"}}, {{uuid "eid"}}, 60);
	insert into escalation_policy_actions (escalation_policy_step_id, user_id) 
	values
		({{uuid "esid"}}, {{uuid "u1"}}),
		({{uuid "esid"}}, {{uuid "u2"}}),
		({{uuid "esid"}}, {{uuid "u3"}});

	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');

	insert into alerts (service_id, description) 
	values
		({{uuid "sid"}}, 'testing');
	`

	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	h.Twilio(t).Device(h.Phone("1")).ExpectSMS("testing")
	h.Twilio(t).Device(h.Phone("2")).ExpectSMS("testing")
	h.Twilio(t).Device(h.Phone("3")).ExpectSMS("testing")
}
