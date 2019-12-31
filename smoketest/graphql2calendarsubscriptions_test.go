package smoketest

import (
	"encoding/json"
	"fmt"
	"github.com/target/goalert/smoketest/harness"
	"testing"
)

// TestGraphQL2Users tests most operations on calendar subscriptions API via GraphQL2 endpoint
func TestGraphQL2CalendarSubscriptions(t *testing.T) {
	t.Parallel()

	sql := `
		insert into users (id, name, email, role) 
		values ({{uuid "user"}}, 'bob', 'joe', 'admin');

		insert into calendar_subscriptions (id, name, user_id)
		values ({{uuid "cs1"}}, 'test1', {{uuid "user"}});
		insert into calendar_subscriptions (id, name, user_id)
		values ({{uuid "cs2"}}, 'test2', {{uuid "user"}});
		insert into calendar_subscriptions (id, name, user_id)
		values ({{uuid "cs3"}}, 'test3', {{uuid "user"}});
	`

	h := harness.NewHarness(t, sql, "calendar-subscriptions")
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

	// find one query
	var cs struct {
		ID         string
		Name       string
		UserID     string
		LastAccess string
		Disabled   bool
	}

	doQL(t, fmt.Sprintf(`
		query {
			calendarSubscription(id: "%s") {
				id
				name
				userID
				lastAccess
				disabled
			}	
		}
	`, h.UUID("cs1")), &cs)

	// find many query
	var subs struct {
		Nodes struct {
			ID         string
			Name       string
			UserID     string
			LastAccess string
			Disabled   bool
		}
	}

	doQL(t, fmt.Sprintf(`
		query {
			calendarSubscriptions(input: {
				first: 3
			}) {
				nodes {
					id
					name
					userID
					lastAccess
					disabled
				}
			}	
		}
	`), &subs)

	// create
	doQL(t, fmt.Sprintf(`
		mutation {
		  createCalendarSubscription(input: {
			name: "%s"
		  })
		}
	`, "Name 1"), nil)

	// update
	doQL(t, fmt.Sprintf(`
		mutation {
		  updateCalendarSubscription(input: {
			id: "%s"
			name: "%s"
			disabled: %v
		  })
		}
	`, h.UUID("cs2"), "Name 2", true), nil)

	// delete
	doQL(t, fmt.Sprintf(`
		mutation {
			deleteAll(input: [{
				id: "%s"
				type: calendarSubscription
			}])
		}
	`, h.UUID("cs3")), nil)
}
