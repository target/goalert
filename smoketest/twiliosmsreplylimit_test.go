package smoketest

import (
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

// TestTwilioSMSReplyLimit checks for a limit on passive replies (no action was taken due to error or already closed, etc).
func TestTwilioSMSReplyLimit(t *testing.T) {
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
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');;

`

	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))

	for i := 0; i < 10; i++ {
		d1.SendSMS("nonsense")
	}

	// only expect 5 replies before we hit the limit
	for i := 0; i < 5; i++ {
		d1.ExpectSMS("sorry")
	}

	h.CreateAlert(h.UUID("sid"), "test1")
	d1.ExpectSMS("test1", "1c", "1a").
		ThenReply("1a").
		ThenExpect("Acknowledged", "#1")

		// should reset after ack
	d1.SendSMS("nonsense")
	d1.ExpectSMS("sorry")

}
