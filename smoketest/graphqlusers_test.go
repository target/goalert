package smoketest

import (
	"encoding/json"
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

// TestGraphQLUsers tests that listing users works properly.
func TestGraphQLUsers(t *testing.T) {
	t.Parallel()

	const sql = `
	insert into users (id, name, email, role)
	values
		({{uuid "u1"}}, 'bob', 'joe', 'user'),
		({{uuid "u2"}}, 'ben', 'josh', 'user');
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

	var res struct {
		Users struct {
			Nodes []struct {
				ID string
			}
		}
	}

	doQL(`
		query {
			users(first: 100) {
				nodes {
					id
				}
			}
		}
	`, &res)
	if len(res.Users.Nodes) != 3 {
		// 3 because the 'GraphQL User' will be implicitly added.
		t.Errorf("got %d users; want 3", len(res.Users.Nodes))
	}
}
