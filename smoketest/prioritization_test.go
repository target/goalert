package smoketest

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

// TestPrioritization tests that notifications for new users/alerts get
// priority over existing ones.
func TestPrioritization(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "u1"}}, 'bob', 'joe');

	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "cm1"}}, {{uuid "u1"}}, 'personal', 'SMS', {{phone "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "u1"}}, {{uuid "cm1"}}, 0);

	insert into escalation_policies (id, name) 
	values
		({{uuid "eid"}}, 'esc policy');

	insert into escalation_policy_steps (id, escalation_policy_id) 
	values
		({{uuid "esid"}}, {{uuid "eid"}});
	insert into escalation_policy_actions (escalation_policy_step_id, user_id) 
	values 
		({{uuid "esid"}}, {{uuid "u1"}});

	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "s1"}}, {{uuid "eid"}}, 'service1'),
		({{uuid "s2"}}, {{uuid "eid"}}, 'service2');
`

	buf := bytes.NewBufferString(`
		insert into alerts (service_id, description)
		values
	`)
	for i := 0; i < 300; i++ {
		if i > 0 {
			buf.WriteString(",\n")
		}
		buf.WriteString(`({{uuid "s1"}}, 'service-1-alert-` + strconv.Itoa(i) + `')`)
	}
	buf.WriteString(";")
	sql += buf.String()

	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	tw := h.Twilio()
	d1 := tw.Device(h.Phone("1"))

	d1.IgnoreUnexpectedSMS("service-1-alert")

	d1.ExpectSMS("service-1-alert")
	tw.WaitAndAssert()

	d1.ExpectSMS("service-1-alert")
	tw.WaitAndAssert()

	h.CreateAlert(h.UUID("s2"), "service-2-alert")

	d1.ExpectSMS("service-2-alert")
	tw.WaitAndAssert()

	d1.ExpectSMS("service-1-alert")
	tw.WaitAndAssert()
}
