package smoke

import (
	"testing"

	"github.com/target/goalert/test/smoke/harness"
)

// TestSMTPAlerts tests that GoAlert responds and
// processes incoming email messages appropriately.
func TestSMTPAlerts(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "user"}}, 'bob', 'joe');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});

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

	insert into integration_keys (id, type, service_id, name) 
	values
		({{uuid "intkey"}}, 'email', {{uuid "sid"}}, 'intkey');
`
	h := harness.NewHarness(t, sql, "trigger-config-sync")
	defer h.Close()

	h.SendMail("foo@example.com", h.UUID("intkey")+"@"+h.Config().EmailIngressDomain(), "test alert", "details")

	h.Twilio(t).Device(h.Phone("1")).ExpectSMS("test alert")
}
