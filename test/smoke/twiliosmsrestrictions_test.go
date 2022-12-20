package smoke

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/test/smoke/harness"
)

// TestTwilioSMSRestrictions checks for restrictions like 1-way and disabling of URLs in messages to certain country codes.
func TestTwilioSMSRestrictions(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email, role) 
	values 
		({{uuid "user"}}, 'bob', 'joe', 'user');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phoneCC "+86" "1"}}),
		({{uuid "cm2"}}, {{uuid "user"}}, 'personal2', 'SMS', {{phoneCC "+1" "1"}}),
		({{uuid "cm3"}}, {{uuid "user"}}, 'personal3', 'SMS', {{phoneCC "+91" "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "user"}}, {{uuid "cm1"}}, 0),
		({{uuid "user"}}, {{uuid "cm2"}}, 0),
		({{uuid "user"}}, {{uuid "cm3"}}, 0);

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

	h.CreateAlert(h.UUID("sid"), "testing")

	tw := h.Twilio(t)

	// US number supports URLs and 2-way SMS
	sms := tw.Device(h.PhoneCC("+1", "1")).ExpectSMS("testing")
	assert.Contains(t, sms.Text(), "ack")
	assert.Contains(t, sms.Text(), "http")

	// CN number does not support URLs or 2-way SMS
	sms = tw.Device(h.PhoneCC("+86", "1")).ExpectSMS("testing")
	assert.NotContains(t, sms.Text(), "ack")
	assert.NotContains(t, sms.Text(), "http")

	// IN supports URLs but not 2-way SMS
	sms = tw.Device(h.PhoneCC("+91", "1")).ExpectSMS("testing")
	assert.NotContains(t, sms.Text(), "ack")
	assert.Contains(t, sms.Text(), "http")
}
