package smoke

import (
	"fmt"
	"testing"

	"github.com/target/goalert/test/smoke/harness"
)

// TestGraphQLServiceLabels tests that labels for services can be created
// (currently only be created directly through db and not via GraphQL layer),
// edited and deleted.

func TestGraphQLServiceLabels(t *testing.T) {
	t.Parallel()

	// Insert initial one label into db
	const sql = `
	insert into escalation_policies (id, name) 
	values
		({{uuid "eid"}}, 'esc policy');
	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');

	insert into labels (tgt_service_id, key, value) 
	values
		({{uuid "sid"}}, 'foo/bar', 'testvalue');
`

	h := harness.NewHarness(t, sql, "labels-switchover-trigger")
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

	// Edit label
	doQL(fmt.Sprintf(`
		mutation {
			setLabel(input:{ target: {type: service , id: "%s"}, key: "%s", value: "%s" }) 
		}
	`, h.UUID("sid"), "foo/bar", "editedvalue"))

	// Delete label
	doQL(fmt.Sprintf(`
		mutation {
			setLabel(input:{ target: {type: service , id: "%s"}, key: "%s", value: "%s" }) 
		}
	`, h.UUID("sid"), "foo/bar", ""))
}
