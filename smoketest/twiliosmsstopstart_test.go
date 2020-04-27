package smoketest

import (
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

// TestTwilioSMSStopStart checks that SMS STOP and START messages are processed.
func TestTwilioSMSStopStart(t *testing.T) {
	t.Parallel()

	sql := `
		insert into users (id, name, email) 
		values 
			({{uuid "user"}}, 'bob', 'joe');
		insert into user_contact_methods (id, user_id, name, type, value) 
		values
			({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}}),
			({{uuid "cm2"}}, {{uuid "user"}}, 'personal', 'VOICE', {{phone "1"}});

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

		insert into alerts (id, service_id, description) 
		values
			(1234, {{uuid "sid"}}, 'testing');
	`

	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	// disable SMS
	h.Twilio().Device(h.Phone("1")).ExpectSMS("testing").ThenReply("stop")
	h.Twilio().Device(h.Phone("1")).ExpectVoice("testing")
	h.Twilio().WaitAndAssert()

	// trigger update - only VOICE should still be enabled
	h.Escalate(1234, 0)
	h.Twilio().Device(h.Phone("1")).ExpectVoice("testing")
	h.Twilio().WaitAndAssert()

	// re-enable SMS
	h.Twilio().Device(h.Phone("1")).SendSMS("start")

	// trigger update - VOICE and SMS should be enabled
	h.Escalate(1234, 0)
	h.Twilio().Device(h.Phone("1")).ExpectSMS("testing")
	h.Twilio().Device(h.Phone("1")).ExpectVoice("testing")
}
