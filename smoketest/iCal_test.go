package smoketest

import (
	"net/http"
	"net/url"
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

	// cfg := h.Config()
	// todo: get Valarm??

	v := make(url.Values)
	token := "some-token"
	v.Set("token", token)

	resp, err := http.PostForm(h.URL()+"/api/v2/calendar", v)
	assert.Nil(t, err)
	if !assert.Equal(t, 200, resp.StatusCode, "serve iCalendar") {
		return
	}

}