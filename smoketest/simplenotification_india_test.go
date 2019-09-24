package smoketest

import (
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

// TestSimpleNotifications_India tests that SMS and Voice goes out international country codes for
// 1 alert -> service -> esc -> step -> user. 2 rules (1 of each) immediately.
// Since with a recent PR, GoAlert accepts all international phone numbers, the negative test for '+222' has been adjusted accordingly.
func TestSimpleNotifications_India(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "user"}}, 'bob', 'joe');
	insert into user_contact_methods (id, user_id, name, type, value)
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phoneCC "+91" "1"}}),
		({{uuid "cm2"}}, {{uuid "user"}}, 'personal', 'VOICE', {{phoneCC "+91" "1"}}),
		({{uuid "cm3"}}, {{uuid "user"}}, 'personal2', 'SMS', {{phoneCC "+222" "1"}}),
		({{uuid "cm4"}}, {{uuid "user"}}, 'personal2', 'VOICE', {{phoneCC "+222" "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "user"}}, {{uuid "cm1"}}, 0),
		({{uuid "user"}}, {{uuid "cm2"}}, 0),
		({{uuid "user"}}, {{uuid "cm3"}}, 0),
		({{uuid "user"}}, {{uuid "cm4"}}, 0);

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
	h := harness.NewStoppedHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	h.Start()

	d1 := h.Twilio().Device(h.PhoneCC("+91", "1"))
	d2 := h.Twilio().Device(h.PhoneCC("+222", "1"))

	d1.ExpectSMS("testing")
	d1.ExpectVoice("testing")
	d2.ExpectSMS("testing")
	d2.ExpectVoice("testing")

}
