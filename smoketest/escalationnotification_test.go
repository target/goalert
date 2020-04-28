package smoketest

import (
	"testing"
	"time"

	"github.com/target/goalert/smoketest/harness"
)

// TestEscalationNotification ensures that notification rules
// don't repeat during an escalation step, and continue to completion.
func TestEscalationNotification(t *testing.T) {
	t.Parallel()
	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "uid"}}, 'bob', 'joe');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "c1"}}, {{uuid "uid"}}, 'personal', 'SMS', {{phone "1"}}),
		({{uuid "c2"}}, {{uuid "uid"}}, 'personal', 'VOICE', {{phone "2"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "uid"}}, {{uuid "c1"}}, 0),
		({{uuid "uid"}}, {{uuid "c2"}}, 0),
		({{uuid "uid"}}, {{uuid "c1"}}, 30);

	insert into escalation_policies (id, name, repeat) 
	values 
		({{uuid "eid"}}, 'esc policy', -1);
	insert into escalation_policy_steps (id, escalation_policy_id, delay) 
	values 
		({{uuid "esid"}}, {{uuid "eid"}}, 60);
	insert into escalation_policy_actions (escalation_policy_step_id, user_id) 
	values 
		({{uuid "esid"}}, {{uuid "uid"}});

	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');

	insert into alerts (service_id, description) 
	values
		({{uuid "sid"}}, 'testing');
`

	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))
	d2 := tw.Device(h.Phone("2"))

	d1.ExpectSMS("testing")
	d2.ExpectVoice("testing")
	tw.WaitAndAssert()

	h.Escalate(1, 0) // results in the start of a 2nd cycle

	d1.ExpectSMS("testing")
	d2.ExpectVoice("testing")
	tw.WaitAndAssert()

	h.FastForward(30 * time.Minute) // ensure both rules have elapsed

	// 1 sms from the first step, 1 from the escalated one
	d1.ExpectSMS("testing")
	d1.ExpectSMS("testing")
	tw.WaitAndAssert()

	h.Escalate(1, 1)
	d1.ExpectSMS("testing")
	d2.ExpectVoice("testing")
	tw.WaitAndAssert()

	h.FastForward(30 * time.Minute)
	d1.ExpectSMS("testing")
}
