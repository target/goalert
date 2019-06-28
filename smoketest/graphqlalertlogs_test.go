package smoketest

import (
	"encoding/json"
	"fmt"
	"github.com/target/goalert/smoketest/harness"
	"testing"
)

// TestGraphQLAlertLogs tests that logs are properly generated for an alert.
func TestGraphQLAlertLogs(t *testing.T) {
	t.Parallel()

	const sql = `
	insert into users (id, name, email, role) 
	values 
		({{uuid "user"}}, 'bob', 'joe', 'user');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "user"}}, {{uuid "cm1"}}, 0),
		({{uuid "user"}}, {{uuid "cm1"}}, 1);
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

	doQL := func(query string, res interface{}) {
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

	sid := h.UUID("sid")
	h.CreateAlert(sid, "alert1")

	// Acknowledging alert
	doQL(fmt.Sprintf(`
		mutation {
			updateAlerts(input: {
				alertIDs: [%d],
				newStatus: StatusAcknowledged,
			}){alertID}
		}
	`, 1), nil)

	// Escalating alert
	doQL(fmt.Sprintf(`
		mutation {
			escalateAlerts(input: [%d],
			){alertID}
		}
	`, 1), nil)

	// Closing alert
	doQL(fmt.Sprintf(`
		mutation {
			updateAlerts(input: {
				alertIDs: [%d],
				newStatus: StatusClosed,
			}){alertID}
		}
	`, 1), nil)

	var logs struct {
		Alert struct {
			RecentEvents struct {
				Nodes []struct {
					Message string `json:"message"`
				} `json:"nodes"`
			} `json:"recentEvents"`
		} `json:"alert"`
	}

	// Verifying log entries exist
	doQL(fmt.Sprintf(`
		query {
  			alert(id: %d) {
    			recentEvents(input: {}) {
					nodes {
						message
					}
    			}
  			}
		}
	`, 1), &logs)

	if len(logs.Alert.RecentEvents.Nodes) < 4 {
		t.Fatalf("ERROR: retrieved length of log entries=%d; want at least %d", len(logs.Alert.RecentEvents.Nodes), 4)
	}
}
