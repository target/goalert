package smoketest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
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

	insert into user_favorites (id, user_id, tgt_service_id)
	values (9999, {{uuid "user"}}, {{uuid "sid2"}});

	insert into alerts (id, service_id, summary, status, dedup_key)
	values
		(1, {{uuid "sid"}}, 'testing notified', 'triggered', 'dedupenotified'),
		(2, {{uuid "sid2"}}, 'testing favorite', 'triggered', 'dedupefavorited');
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
			first: 2
			includeNotified: false
			favoritesOnly: true
		}) {
			nodes {
				id
			}
		}
	}`, &alerts1)

	var e struct {
		Alerts struct {
			Nodes []struct {
				ID string
			}
		}
	}
	//emptySlice := make([]string, 0)
	// 	 e [2]int
	// e := []struct { ID string }
	// e[0] = 2
	// }

	assert.Equal(t, e.Alerts.Nodes, alerts1.Alerts.Nodes)

	// test:
	// includeNotified: true
	// favoritesOnly: true
	// output: 2 alerts (1 from favorited, 1 from notified)

	// QUESTIONS
	// Should we be using "fmt.Sprintf" When writing the queries. Why or why not?
	// Should Alerts be delared at the begining of class like with graphqlmultiplealerts

	doQL(t, h, `query {
		alerts(input: {
			first: 2
			includeNotified: true
			favoritesOnly: true
		}) {
			nodes {
				id
			}
		}
	}`, &alerts2)
}

// `query {
// 	users(first: 100) {
// 		nodes {
// 			id
// 		}
// 	}
// }
// `
