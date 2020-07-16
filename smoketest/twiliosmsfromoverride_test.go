package smoketest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/smoketest/harness"
)

// TestTwilioSMSFromOverride checks that the FromNumber of an SMS can be overridden per-carrier.
func TestTwilioSMSFromOverride(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email, role) 
	values 
		({{uuid "user"}}, 'bob', 'joe', 'user');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}}),
		({{uuid "cm2"}}, {{uuid "user"}}, 'personal2', 'SMS', {{phone "2"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "user"}}, {{uuid "cm1"}}, 0),
		({{uuid "user"}}, {{uuid "cm2"}}, 0);

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

	from := h.TwilioNumber("")
	altFrom := h.TwilioNumber("alt1")
	h.SetConfigValue("Twilio.SMSFromNumberOverride", "foobar="+altFrom)
	h.SetConfigValue("Twilio.SMSCarrierLookup", "true")
	h.SetCarrierName(h.Phone("2"), "foobar")

	h.CreateAlert(h.UUID("sid"), "testing")

	tw := h.Twilio(t)
	sms1 := tw.Device(h.Phone("1")).ExpectSMS("testing")
	assert.Equal(t, from, sms1.From(), "first message from number")

	sms2 := tw.Device(h.Phone("2")).ExpectSMS("testing")
	assert.Equal(t, altFrom, sms2.From(), "overridden carrier from number")

}
