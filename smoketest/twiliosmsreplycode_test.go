package smoketest

import (
	"github.com/target/goalert/smoketest/harness"
	"testing"
)

// TestTwilioSMSReplyCode checks that reply codes work properly.
func TestTwilioSMSReplyCode(t *testing.T) {
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
		({{uuid "user"}}, {{uuid "cm1"}}, 0),
		({{uuid "user"}}, {{uuid "cm1"}}, 1);

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

	tw := h.Twilio()
	d1 := tw.Device(h.Phone("1"))

	h.CreateAlert(h.UUID("sid"), "test1")
	d1.ExpectSMS("test1", "1c", "1a").ThenReply("1a")
	d1.ExpectSMS("Acknowledged", "#1")

	tw.WaitAndAssert()

	h.CreateAlert(h.UUID("sid"), "test2")
	d1.ExpectSMS("test2", "2c", "2a").ThenReply("1a") // ack again
	d1.ExpectSMS("already", "ack").ThenReply("'1c'")  // then close
	d1.ExpectSMS("Closed", "#1")

	tw.WaitAndAssert()

	h.CreateAlert(h.UUID("sid"), "test3")
	d1.ExpectSMS("test3", "1c", "1a").ThenReply("1 a") // 1 was re-used for alert #3

	tw.WaitAndAssert()

	d1.ExpectSMS("Ack", "#3")

	tw.WaitAndAssert()

	h.CreateAlert(h.UUID("sid"), "test4")
	d1.ExpectSMS("test4", "3c", "3a").ThenReply("close 4") // old method 'close alertID' still works
	d1.ExpectSMS("Closed", "#4")
}
