package smoketest

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

func TestNotifiedAlerts(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email, role) 
	values ({{uuid "user"}}, 'bob', 'joe', 'user');

	insert into user_contact_methods (id, user_id, name, type, value) 
	values ({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values ({{uuid "user"}}, {{uuid "cm1"}}, 0);

	insert into escalation_policies (id, name) 
	values
		({{uuid "eid"}}, 'esc policy'),
		({{uuid "eid2"}}, 'esc policy 2');

	insert into escalation_policy_steps (id, escalation_policy_id) 
	values ({{uuid "esid"}}, {{uuid "eid"}});
 

	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service'),
		({{uuid "sid2"}}, {{uuid "eid2"}}, 'service 2');
	`

	h := harness.NewHarness(t, sql, "add-no-notification-alert-log")
	defer h.Close()

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

	doQL(t, h, fmt.Sprintf(`
		mutation{
			setFavorite(input:{
				target:{
					id: "%s"
					type: service
				}
				favorite: true
			})
		}
	`, h.UUID("sid2")), nil)

	var s struct {
		Service struct {
			IsFavorite bool
		}
	}

	doQL(t, h, fmt.Sprintf(`
		query {
			service(id: "%s") { 
				isFavorite 
			}	
		}
	`, h.UUID("sid2")), &s)

	if s.Service.IsFavorite != true {
		t.Fatalf("ERROR: ServiceID %s IsUserFavorite=%t; want true", h.UUID("sid2"), s.Service.IsFavorite)
	}

	h.CreateAlert(h.UUID("sid"), "alert1")
	h.CreateAlert(h.UUID("sid2"), "alert2")

	type Alerts struct {
		Alerts struct {
			Nodes []struct {
				ID string
			}
		}
	}

	var alerts1, alerts2 Alerts

	// notes
	// - "user" is assigned to "sid"
	// - "sid" is not favorited by "user"
	// - it has 1 alert that "user" should be notified for

	// - "sid2" created and has no assignments
	// - it is favorited by "user"
	// - it has 1 alert

	// test:
	// includeNotified: false
	// favoritesOnly: true
	// output: 1 alert (the favorited one)
	doQL(t, h, `query {
		alerts(input: {
			includeNotified: false
			favoritesOnly: true
		}) {
			nodes {
				id
			}
		}
	}`, &alerts1)

	if len(alerts1.Alerts.Nodes) != 1 {
		t.Errorf("got %d alerts; want 1", len(alerts1.Alerts.Nodes))
	}

	// test:
	// includeNotified: true // service1
	// favoritesOnly: true // service2
	// output: 2 alerts (1 from favorited, 1 from notified)

	doQL(t, h, `query {
			alerts(input: {
				includeNotified: true
				favoritesOnly: true
			}) {
				nodes {
					id
				}
			}
		}`, &alerts2)

	if len(alerts2.Alerts.Nodes) != 2 {
		t.Errorf("got %d alerts; want 2", len(alerts2.Alerts.Nodes))
	}

}
