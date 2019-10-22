package smoketest

import (
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

// TestManualEscalation ensures that second step notifications are sent out when an acknowledged alert is manually escalated.
// When an acknowledged alert is manually escalated, it should escalate and go back to the 'unacknowleged' state.
// TestManualEscalation should create an alert in the acknowledged/active state, with a 2+ step EP, then trigger an escalation. Assert that the second step notifications are sent

func TestManualEscalation(t *testing.T) {
	t.Parallel()
	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "uid"}}, 'bob', 'joe'),
		({{uuid "uid2"}}, 'jane', 'xyz');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "c1"}}, {{uuid "uid"}}, 'personal', 'SMS', {{phone "1"}}),
		({{uuid "c2"}}, {{uuid "uid2"}}, 'personal', 'SMS', {{phone "2"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "uid"}}, {{uuid "c1"}}, 0),
		({{uuid "uid2"}}, {{uuid "c2"}}, 0);
	
	insert into escalation_policies (id, name, repeat) 
	values 
		({{uuid "eid"}}, 'esc policy', -1);
	insert into escalation_policy_steps (id, escalation_policy_id, delay) 
	values 
		({{uuid "esid1"}}, {{uuid "eid"}}, 60),
		({{uuid "esid2"}}, {{uuid "eid"}}, 60);

	insert into escalation_policy_actions (escalation_policy_step_id, user_id) 
	values 
		({{uuid "esid1"}}, {{uuid "uid"}}),
		({{uuid "esid2"}}, {{uuid "uid2"}});

	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');

	insert into alerts (service_id, description, status) 
	values
		({{uuid "sid"}}, 'testing', 'active');
`

	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	h.Twilio().WaitAndAssert() // phone 2 should not get SMS before escalating
	h.Escalate(1, 0)

	h.Twilio().Device(h.Phone("2")).ExpectSMS("testing")
}
