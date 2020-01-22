package smoketest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/smoketest/harness"
)

func TestCalendarSubscription(t *testing.T) {
	t.Parallel()

	const sql = `
	insert into users (id, name, email)
	values
		({{uuid "user"}}, 'bob', 'joe');
	insert into schedules (id, name, time_zone, description) 
	values
		({{uuid "schedId"}},'sched', 'America/Chicago', 'test description here');
	`
	h := harness.NewHarness(t, sql, "calendar-subscriptions")
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

	var cs struct{ CreateUserCalendarSubscription struct{ URL string } }

	doQL(fmt.Sprintf(`
		mutation {
			createUserCalendarSubscription (input: {
				name: "%s",
				reminderMinutes: [%d]
				scheduleID: "%s",
			}) {
				url
			}
		}
	`, "foobar", 5, h.UUID("schedId")), &cs)

	u, err := url.Parse(cs.CreateUserCalendarSubscription.URL)
	assert.Nil(t, err)
	assert.Contains(t, u.Path, "/api/v2/calendar")

	resp, err := http.Get(cs.CreateUserCalendarSubscription.URL)
	assert.Nil(t, err)
	if !assert.Equal(t, 200, resp.StatusCode, "serve iCalendar") {
		return
	}
}
