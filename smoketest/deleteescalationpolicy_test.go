package smoketest

import (
	"fmt"
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

// TestDeleteEscalationPolicy tests that it is possible to delete an escalation policy
func TestDeleteEscalationPolicy(t *testing.T) {
	t.Parallel()

	const sql = `
	insert into escalation_policies (id, name, description)
	values
		({{uuid "ep1"}}, 'test', 'test');
	
	insert into escalation_policy_steps (id, escalation_policy_id)
	values
		({{uuid ""}}, {{uuid "ep1"}}),
		({{uuid ""}}, {{uuid "ep1"}});
`

	h := harness.NewHarness(t, sql, "heartbeats")
	defer h.Close()

	doQL := func(query string) {
		g := h.GraphQLQuery2(query)
		for _, err := range g.Errors {
			t.Error("GraphQL Error:", err.Message)
		}
		if len(g.Errors) > 0 {
			t.Fatal("errors returned from GraphQL")
		}
		t.Log("Response:", string(g.Data))
	}

	doQL(fmt.Sprintf(`
		mutation {
			deleteAll(input:{id: "%s", type: escalationPolicy})
		}
	`, h.UUID("ep1")))
}
