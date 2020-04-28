package smoketest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/smoketest/harness"
)

// TestTwilioURL_SMS checks that the GoAlert link in the SMS is correct
// when General.ShortURL or General.DisableSMSLinks are enabled, as well
// as with the default URL
func TestTwilioURL_SMS(t *testing.T) {
	t.Parallel()

	sql := `
		insert into users (id, role, name, email) 
		values 
			({{uuid "user"}}, 'user', 'bob', 'joe');
		insert into user_contact_methods (id, user_id, name, type, value) 
		values
			({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});
		insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
		values
			({{uuid "user"}}, {{uuid "cm1"}}, 0);
		update users set alert_status_log_contact_method_id = {{uuid "cm1"}}
		where id = {{uuid "user"}};

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
			({{uuid "sid"}}, {{uuid "eid"}}, 'My Service');
	`

	const shortURL = "http://sho.rt"

	t.Run("default URL in sms body", func(t *testing.T) {
		t.Parallel()
		h := harness.NewHarness(t, sql, "message-bundles")
		defer h.Close()

		tw := h.Twilio(t)
		d1 := tw.Device(h.Phone("1"))

		longURL := h.URL()

		h.CreateAlert(h.UUID("sid"), "test")
		d1.ExpectSMS("test", longURL)

		tw.WaitAndAssert()
	})

	t.Run("General.ShortURL in sms body", func(t *testing.T) {
		t.Parallel()
		h := harness.NewHarness(t, sql, "message-bundles")
		defer h.Close()

		tw := h.Twilio(t)
		d1 := tw.Device(h.Phone("1"))

		h.SetConfigValue("General.ShortURL", shortURL)

		h.CreateAlert(h.UUID("sid"), "test")
		d1.ExpectSMS("test", shortURL)

		tw.WaitAndAssert()
	})

	t.Run("General.DisableSMSLinks with General.ShortURL set", func(t *testing.T) {
		t.Parallel()
		h := harness.NewHarness(t, sql, "message-bundles")
		defer h.Close()

		tw := h.Twilio(t)
		d1 := tw.Device(h.Phone("1"))

		h.SetConfigValue("General.ShortURL", shortURL)
		h.SetConfigValue("General.DisableSMSLinks", "true")

		h.CreateAlert(h.UUID("sid"), "test")
		smsMsg := d1.ExpectSMS("test")
		tw.WaitAndAssert()
		assert.NotContains(t, smsMsg.Body(), "http")
	})

	t.Run("General.DisableSMSLinks using default URL", func(t *testing.T) {
		t.Parallel()
		h := harness.NewHarness(t, sql, "message-bundles")
		defer h.Close()

		tw := h.Twilio(t)
		d1 := tw.Device(h.Phone("1"))

		longURL := h.URL()
		h.SetConfigValue("General.DisableSMSLinks", "true")

		h.CreateAlert(h.UUID("sid"), "test")
		smsMsg := d1.ExpectSMS("test")
		tw.WaitAndAssert()
		assert.NotContains(t, smsMsg.Body(), longURL)
		assert.NotContains(t, smsMsg.Body(), "http")
	})
}
