package smoketest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/smoketest/harness"
)

func TestICal(t *testing.T) {
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

	// get token
	s := strings.Split(cs.CreateUserCalendarSubscription.URL, "token=")
	v := make(url.Values)
	v.Set("token", s[1])

	resp, err := http.PostForm(h.URL()+"/api/v2/calendar", v)
	assert.Nil(t, err)
	if !assert.Equal(t, 200, resp.StatusCode, "serve iCalendar") {
		return
	}

}