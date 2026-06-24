package smoke

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

var (

	// example: 2025-12-13T16:06:38.918293-06:00
	isoRx     = regexp.MustCompile(`"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(Z|[-+]\d{2}:\d{2})"`)
	urlHostRx = regexp.MustCompile(`"http://[^/]+`)
)

func TestCalendarSubscriptionJSON(t *testing.T) {
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

	const mut = `
		mutation {
			createUserCalendarSubscription (input: {
				name: "%s",
				reminderMinutes: [%d]
				scheduleID: "%s",
			}) {
				url
			}
		}
	`

	// create subscription
	doQL(fmt.Sprintf(mut, "foobar", 5, h.UUID("schedId")), &cs)

	u, err := url.Parse(cs.CreateUserCalendarSubscription.URL)
	assert.NoError(t, err)
	assert.Contains(t, u.Path, "/api/v2/calendar")

	resp, err := http.Get(cs.CreateUserCalendarSubscription.URL)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode, "serve iCalendar")

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, "BEGIN:VCALENDAR\r\nPRODID:-//GoAlert//dev//EN\r\nVERSION:2.0\r\nCALSCALE:GREGORIAN\r\nMETHOD:PUBLISH\r\nEND:VCALENDAR\r\n", string(data))

	req, err := http.NewRequest("GET", cs.CreateUserCalendarSubscription.URL, nil)
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode, "serve JSON")
	data, err = io.ReadAll(resp.Body)
	require.NoError(t, err)

	data = isoRx.ReplaceAll(data, []byte(`"2021-01-01T00:00:00Z"`))
	data = urlHostRx.ReplaceAll(data, []byte(`"http://TEST_HOST`))

	expected := fmt.Sprintf(`
{
"AppName": "GoAlert",
"AppVersion": "dev",
"Start": "2021-01-01T00:00:00Z",
"End": "2021-01-01T00:00:00Z",
"ScheduleID": "%s",
"ScheduleName": "sched",
"ScheduleURL": "http://TEST_HOST/schedules/%s",
"Shifts":[],
"Type": "calendar-subscription/v1"
}
`, h.UUID("schedId"), h.UUID("schedId"))

	assert.JSONEq(t, expected, string(data))
}
