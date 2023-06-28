package smoke

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/target/goalert/test/smoke/harness"
)

// TestGraphQLMultipleAlerts tests that all steps up to, and including, generating
// alerts and updating their statuses via GraphQL.
func TestGraphQLMultipleAlerts(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email, role) 
	values 
		({{uuid "user"}}, 'bob', 'joe', 'user');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "user"}}, {{uuid "cm1"}}, 0),
		({{uuid "user"}}, {{uuid "cm1"}}, 30);

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
`

	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	phone := h.Phone("1")
	sid := h.UUID("sid")

	// Creating alerts
	h.CreateAlert(sid, "alert1")
	h.CreateAlert(sid, "alert2")

	// Expect 2 SMS for 2 unacknowledged alerts
	h.Twilio(t).Device(phone).ExpectSMS("alert1")
	h.Twilio(t).Device(phone).ExpectSMS("alert2")

	h.CreateAlert(sid, "alert3")

	// GraphQL2 section starts
	doQL2 := func(query string, res interface{}) {
		g := h.GraphQLQuery2(query)
		for _, err := range g.Errors {
			t.Error("GraphQL Error:", err.Message)
		}
		if len(g.Errors) > 0 {
			t.Fatal("errors returned from GraphQL")
		}
		t.Log("Response:", string(g.Data))
		if res == nil {
			return
		}
		err := json.Unmarshal(g.Data, &res)
		if err != nil {
			t.Fatal("failed to parse response:", err)
		}
	}

	h.Twilio(t).Device(phone).ExpectSMS("alert3")

	// Acknowledging alert #3
	doQL2(fmt.Sprintf(`
		mutation {
			updateAlerts(input: {
				alertIDs: [%d],
				newStatus: StatusAcknowledged,
			}){alertID}
		}
	`, 3), nil)

	h.FastForward(30 * time.Minute) // notification rule
	// Expect 2 SMS for 2 unacknowledged alerts
	h.Twilio(t).Device(phone).ExpectSMS("alert1")
	h.Twilio(t).Device(phone).ExpectSMS("alert2")

	h.FastForward(30 * time.Minute)

	// Escalating multiple (3) alerts
	esc := func(id int) {
		t.Helper()
		doQL2(fmt.Sprintf(`
		mutation {
			escalateAlerts(input: [%d],
			){alertID}
		}
	`, id), nil)
	}
	esc(1)
	esc(2)
	esc(3)

	// Expect 3 SMS for 3 escalated alerts
	h.Twilio(t).Device(phone).ExpectSMS("alert1")
	h.Twilio(t).Device(phone).ExpectSMS("alert2")
	h.Twilio(t).Device(phone).ExpectSMS("alert3")

	// Closing multiple (3) alerts
	doQL2(fmt.Sprintf(`
		mutation {
			updateAlerts(input: {
				alertIDs: [%d, %d, %d],
				newStatus: StatusClosed,
			}){alertID}
		}
	`, 1, 2, 3), nil)

	h.FastForward(1 * time.Minute)
}
