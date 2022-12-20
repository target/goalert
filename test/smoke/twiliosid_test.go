package smoke

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/test/smoke/harness"
)

// TestTwilioSID checks that messages to SMS and Voice numbers work when using a Messaging Service SID for Twilio.FromNumber.
func TestTwilioSID(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email, role) 
	values 
		({{uuid "user"}}, 'bob', 'joe', 'user');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}}),
		({{uuid "cm2"}}, {{uuid "user"}}, 'personal', 'VOICE', {{phone "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "user"}}, {{uuid "cm1"}}, 0),
		({{uuid "user"}}, {{uuid "cm2"}}, 0),
		({{uuid "user"}}, {{uuid "cm1"}}, 30),
		({{uuid "user"}}, {{uuid "cm2"}}, 30);

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
`
	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	h.SetConfigValue("Twilio.MessagingServiceSID", h.TwilioMessagingService())

	h.CreateAlert(h.UUID("sid"), "testing")

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))

	sms := d1.ExpectSMS("testing")

	// should use a messaging service number
	assert.NotEqual(t, h.TwilioNumber(""), sms.From())
	sms.ThenReply("ack 1").
		ThenExpect("acknowledged")

	d1.ExpectCall().Answer().
		ExpectSay("testing").
		Press("6").
		ExpectSay("closed").
		Hangup()

	h.FastForward(time.Hour)
	// no more messages
}
