package smoketest

import (
	"github.com/target/goalert/smoketest/harness"
	"testing"
	"time"
)

// TestTwilioVoiceStop checks that a voice call STOP is processed.
func TestTwilioVoiceStop(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "user"}}, 'bob', 'joe');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'VOICE', {{phone "1"}}),
		({{uuid "cm2"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "user"}}, {{uuid "cm1"}}, 0),
		({{uuid "user"}}, {{uuid "cm1"}}, 1),
		({{uuid "user"}}, {{uuid "cm2"}}, 1);

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
	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	d1 := h.Twilio().Device(h.Phone("1"))

	d1.ExpectVoice("testing").
		ThenPress("1").
		ThenExpect("unenrollment").
		ThenPress("3").
		ThenExpect("goodbye")

	// Should unenroll completely (no voice or SMS)
	h.Twilio().WaitAndAssert()

	h.FastForward(time.Minute)

	h.Delay(time.Second * 15)
	// no more messages, it should have disabled both
}
