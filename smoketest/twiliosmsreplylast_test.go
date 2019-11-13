package smoketest

import (
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

// TestTwilioSMSReplyLast checks that an SMS reply message is processed with no number.
func TestTwilioSMSReplyLast(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email, role) 
	values 
		({{uuid "user"}}, 'bob', 'joe', 'user');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "user"}}, {{uuid "cm1"}}, 0);

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

	insert into alerts (id, service_id, description) 
	values
		(198, {{uuid "sid"}}, 'testing');

`
	check := func(respondWith, expect string) {
		t.Run("check", func(t *testing.T) {
			h := harness.NewHarness(t, sql, "ids-to-uuids")
			defer h.Close()

			tw := h.Twilio()
			d1 := tw.Device(h.Phone("1"))

			d1.ExpectSMS("testing").ThenReply(respondWith)
			d1.ExpectSMS(expect, "198")
		})
	}

	check("ack", "acknowledged")
	check("a", "acknowledged")
	check("close", "closed")
	check("c", "closed")
}
