package smoketest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/smoketest/harness"
)

// TestGraphQLAlertSort verifies that the alerts query sorts results properly.
func TestGraphQLAlertSort(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email, role) 
	values 
		({{uuid "user"}}, 'bob', 'joe', 'user');

	insert into escalation_policies (id, name) 
	values
		({{uuid "eid"}}, 'esc policy');

	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');

	insert into alerts (id, service_id, summary, created_at, status, dedup_key) 
	values
		(1, {{uuid "sid"}}, 'Notified alert', now(), 'active', 'auto:1:foo'),
		(2, {{uuid "sid"}}, 'Favorited alert', now() - '1 days'::interval, 'active', 'auto:1:bar'),
		(3, {{uuid "sid"}}, 'Notified alert', now(), 'closed', NULL),
		(4, {{uuid "sid"}}, 'Notified alert', now() - '3 days'::interval, 'closed', NULL);
	
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

	type node struct {
		ID int `json:",string"`
	}

	type Alerts struct {
		Alerts struct {
			Nodes []node
		}
	}

	var alerts1, alerts2, alerts3 Alerts

	// Expected output based on when alert was created by ID from newest to oldest
	doQL(t, h, `query {
		alerts(input: { sort: dateID }) {
		  nodes {
			id
		  }
		}
	  }`, &alerts1)

	assert.EqualValues(t, []node{{ID: 3}, {ID: 1}, {ID: 2}, {ID: 4}}, alerts1.Alerts.Nodes)

	// Expected output based on the status of the alert by ID
	doQL(t, h, `query {
		alerts(input: { sort: statusID }) {
		  nodes {
			id
		  }
		}
	  }`, &alerts2)

	assert.EqualValues(t, []node{{ID: 2}, {ID: 1}, {ID: 4}, {ID: 3}}, alerts2.Alerts.Nodes)

	// expected output is based on when alert was created by Id from oldest to newest
	doQL(t, h, `query {
		alerts(input: { sort: dateIDReverse }) {
		  nodes {
			id
		  }
		}
	  }`, &alerts3)

	assert.EqualValues(t, []node{{ID: 4}, {ID: 2}, {ID: 1}, {ID: 3}}, alerts3.Alerts.Nodes)

}
