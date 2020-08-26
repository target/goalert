package smoketest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/smoketest/harness"
)

// TestNotifiedAlerts tests that the alerts GraphQL query shows the proper amount of results when flipping between "includeNotified" and "favoritesOnly" query options.
// Service 1: Notified alert, Service 2: Favorited alert
func TestNotifiedAlerts(t *testing.T) {
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
		({{uuid "user"}}, {{uuid "cm1"}}, 0);

	insert into escalation_policies (id, name) 
	values
		({{uuid "eid"}}, 'esc policy'),
		({{uuid "eid2"}}, 'esc policy 2');

	insert into escalation_policy_steps (id, escalation_policy_id) 
	values 
		({{uuid "esid"}}, {{uuid "eid"}});
 
	insert into escalation_policy_actions (id, escalation_policy_step_id, user_id)
	values 
		({{uuid "epa"}}, {{uuid "esid"}}, {{uuid "user"}});

	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service'),
		({{uuid "sid2"}}, {{uuid "eid2"}}, 'service 2');

	insert into alerts (service_id, summary, dedup_key) 
	values
		({{uuid "sid"}}, 'Notified alert', 'auto:1:foo'),
		({{uuid "sid2"}}, 'Favorited alert', 'auto:1:bar');

	insert into user_favorites (user_id, tgt_service_id)
	values
		({{uuid "user"}}, {{uuid "sid2"}});	
	`

	h := harness.NewHarness(t, sql, "add-no-notification-alert-log")
	defer h.Close()

	doQL := func(t *testing.T, h *harness.Harness, query string, res interface{}) {
		t.Helper()
		g := h.GraphQLQueryUserT(t, h.UUID("user"), query)
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

	type Alerts struct {
		Alerts struct {
			Nodes []struct {
				ID string
			}
		}
	}

	var alerts1, alerts2, alerts3 Alerts

	// output: 1 alert (the favorited one)
	doQL(t, h, `query {
		alerts(input: {
			includeNotified: false
			favoritesOnly: true
		}) {
			nodes {
				id
				summary
			}
		}
	}`, &alerts1)

	assert.Len(t, alerts1.Alerts.Nodes, 1, "alerts query")

	// Expect 1 SMS for the created alert
	h.Twilio(t).Device(h.Phone("1")).ExpectSMS("notified")

	// query for favorites & notified: 2 alerts
	doQL(t, h, `query {
			alerts(input: {
				includeNotified: true
				favoritesOnly: true
			}) {
				nodes {
					id
					summary
				}
			}
		}`, &alerts2)

	assert.Len(t, alerts2.Alerts.Nodes, 2, "alerts query")

	// All Services test (favoritesOnly: false)
	// output: 2 alerts
	doQL(t, h, `query {
		alerts(input: {
			includeNotified: true
			favoritesOnly: false
		}) {
			nodes {
				id
				summary
			}
		}
	}`, &alerts3)

	assert.Len(t, alerts3.Alerts.Nodes, 2, "alerts query")
}
