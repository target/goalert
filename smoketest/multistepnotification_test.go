package smoketest

import (
	"github.com/target/goalert/smoketest/harness"
	"testing"
	"time"
)

// TestMultiStepNotifications tests that SMS and Voice goes out for
// 1 alert -> service -> esc -> step -> user. with 3 notification rules (1 of each immediately, sms 1 minute later)
func TestMultiStepNotifications(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "u1"}}, 'bob', 'joe');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "c1"}}, {{uuid "u1"}}, 'personal', 'SMS', {{phone "1"}}),
		({{uuid "c2"}}, {{uuid "u1"}}, 'personal', 'VOICE', {{phone "2"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "u1"}}, {{uuid "c1"}}, 0),
		({{uuid "u1"}}, {{uuid "c2"}}, 0),
		({{uuid "u1"}}, {{uuid "c1"}}, 1);

	insert into escalation_policies (id, name) 
	values 
		({{uuid "e1"}}, 'esc policy');
	insert into escalation_policy_steps (id, escalation_policy_id) 
	values 
		({{uuid "es1"}}, {{uuid "e1"}});
	insert into escalation_policy_actions (escalation_policy_step_id, user_id) 
	values 
		({{uuid "es1"}}, {{uuid "u1"}});

	insert into services (id, escalation_policy_id, name) 
	values
    	({{uuid "s1"}}, {{uuid "e1"}}, 'service');

	insert into alerts (service_id, description) 
	values
    	({{uuid "s1"}}, 'testing');
`
	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	tw := h.Twilio()
	d1 := tw.Device(h.Phone("1"))
	d2 := tw.Device(h.Phone("2"))

	d1.ExpectSMS("testing")
	d2.ExpectVoice("testing")
	tw.WaitAndAssert()

	h.FastForward(time.Minute)
	d1.ExpectSMS("testing")

}
