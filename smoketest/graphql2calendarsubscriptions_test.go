package smoketest

import (
	"encoding/json"
	"fmt"
	"github.com/target/goalert/smoketest/harness"
	"testing"
)

// TestGraphQL2Users tests most operations on calendar subscriptions API via GraphQL2 endpoint.
func TestGraphQL2CalendarSubscriptions(t *testing.T) {
	t.Parallel()

	sql := `
		insert into users (id, name, email, role) 
		values 
			({{uuid "user"}}, 'bob', 'joe', 'admin');
	`

	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	doQL := func(t *testing.T, query string, res interface{}) {
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

	var cs struct {
		ID string
		Name string
		UserID string
		LastAccess string
		Disabled bool
	}

	// find one query

	// find many query

	// create
	doQL(t, fmt.Sprintf(`
		mutation {
		  createCalendarSubscription(input:{
			name: "%s"
		  })
		}
	`, "Name 1"), nil)

	// update

	// delete
}
