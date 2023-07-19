package smoke

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/test/smoke/harness"
)

func TestAlertLogDelay(t *testing.T) {
	t.Parallel()

	doQL := func(t *testing.T, h *harness.Harness, query string, res interface{}) {
		t.Helper()
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

	const query = `
		insert into users (id, name, email) 
		values ({{uuid "user"}}, 'bob', 'joe');

		insert into user_contact_methods (id, user_id, name, type, value, disabled) 
		values ({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}}, false);

		insert into escalation_policies (id, name) 
		values ({{uuid "eid"}}, 'esc policy');

		insert into services (id, escalation_policy_id, name) 
		values ({{uuid "sid"}}, {{uuid "eid"}}, 'service');

		insert into alerts (id, service_id, summary, dedup_key)
		values (10, {{uuid "sid"}}, 'test_summary', 'auto:1:test');

		insert into outgoing_messages (id, message_type, last_status_at, created_at, last_status, alert_id, user_id, service_id, escalation_policy_id, sent_at, contact_method_id)
		values ({{uuid "omid"}}, 'alert_notification', NOW(), NOW() - '5 minutes'::interval, 'delivered', 10, {{uuid "user"}}, {{uuid "sid"}}, {{uuid "eid"}}, NOW(), {{uuid "cm1"}});

		insert into alert_logs (alert_id, event, meta, message)
		values (10, 'notification_sent', '{"MessageID": {{uuidJSON "omid"}}}', '');
	`

	h := harness.NewHarness(t, query, "add-no-notification-alert-log")
	defer h.Close()

	var alertLogs struct {
		Alert struct {
			RecentEvents struct {
				Nodes []struct {
					State struct {
						Details string
					}
				}
			}
		}
	}

	doQL(t, h, fmt.Sprintf(`
		query {
			alert(id: %d) {
				recentEvents(input: { limit: 15 }) {
					nodes {
						message
						state {
							details
						}
					}
				}
			}
		}
	`, 10), &alertLogs)

	assert.Contains(t, "Delivered (after", alertLogs.Alert.RecentEvents.Nodes[0].State.Details)
}
