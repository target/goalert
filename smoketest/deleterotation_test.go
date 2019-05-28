package smoketest

import (
	"fmt"
	"github.com/target/goalert/smoketest/harness"
	"testing"
)

// TestDeleteRotation tests that it is possible to delete a rotation with participants
func TestDeleteRotation(t *testing.T) {
	t.Parallel()

	const sql = `
	insert into users (id, name, email, role)
	values
		({{uuid "u1"}}, 'bob', 'joe', 'user'),
		({{uuid "u2"}}, 'ben', 'josh', 'user');
	
	insert into rotations (id, name, description, type, start_time, time_zone)
	values
		({{uuid "r1"}}, 'test', 'test', 'daily', now(), 'UTC');
	
	insert into rotation_participants (id, rotation_id, user_id, position)
	values
		({{uuid ""}}, {{uuid "r1"}}, {{uuid "u1"}},0),
		({{uuid ""}}, {{uuid "r1"}}, {{uuid "u2"}},1);
`

	h := harness.NewHarness(t, sql, "heartbeats")
	defer h.Close()

	doQL := func(query string) {
		g := h.GraphQLQuery(query)
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
			deleteRotation(input:{id: "%s"}) {
				deleted_id
			}
		}
	`, h.UUID("r1")))
}
