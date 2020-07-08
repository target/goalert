package smoketest

import (
	"encoding/json"
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

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
		({{uuid "user"}}, {{uuid "cm1"}}, 0),
		

	insert into escalation_policies (id, name) 
	values
		({{uuid "eid"}}, 'esc policy'),
		({{uuid "eid2"}}, 'esc policy');

	insert into escalation_policy_steps (id, escalation_policy_id) 
	values
		({{uuid "esid"}}, {{uuid "eid"}});
		

	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service'),
		({{uuid "sid2"}}, {{uuid "eid2"}}, 'service');

		insert into alerts (id, service_id, summary, status, created_at) 
		values
			(1, {{uuid "sid"}}, 'testing', 'unacknowledged', now()),
			(2, {{uuid "sid2"}}, 'testing', 'unacknowledged', now() - '2 days'::interval);
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

	var alerts1, alerts2 struct {
		AlertConnection struct {
			Nodes []struct {
				ID string
			}
		}
	}

	doQL(t, h, "", &alerts1)

}
