package smoketest

import (
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

// TestTwilioVoiceFailure checks that a voice call failure is retried.
func TestTwilioVoiceFailure(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "user"}}, 'bob', 'joe');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'VOICE', {{phone "1"}});

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

	insert into alerts (service_id, description) 
	values
		({{uuid "sid"}}, 'testing');

`
	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	d1 := h.Twilio(t).Device(h.Phone("1"))
	d1.RejectVoice("testing")

	// should not retry (failed calls are perm failure)
	h.Trigger()
}
