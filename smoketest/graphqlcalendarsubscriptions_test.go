package smoketest

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/smoketest/harness"
)

type calSub struct {
	Name            string
	Disabled        bool
	ScheduleID      string
	ReminderMinutes []int
}
type calSubWithID struct {
	ID string
	calSub
}

// TestGraphQLCalendarSubscriptions tests operations on calendar subscriptions API
func TestGraphQLCalendarSubscriptions(t *testing.T) {
	t.Parallel()

	sql := `
		insert into users (id, name, email, role) 
		values ({{uuid "user"}}, 'bob', 'joe', 'admin');

		insert into schedules (id, name, time_zone) 
		values ({{uuid "sched1"}}, 'default', 'America/Chicago');

		insert into user_calendar_subscriptions (id, name, user_id, config, schedule_id)
		values ({{uuid "cs1"}}, 'test1', {{uuid "user"}}, '{ "ReminderMinutes": [2, 4, 8, 16]}', {{uuid "sched1"}});

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

	// create
	var csCreate struct {
		CreateUserCalendarSubscription calSubWithID
	}
	doQL(t, fmt.Sprintf(`
		mutation {
		  createUserCalendarSubscription(input: {
			name: "%s"
			scheduleID: "%s"
			reminderMinutes: [%d]
			disabled: %v
		  }) {
			id
			name
			scheduleID
			reminderMinutes
		  }
		}
	`, "Name 1", h.UUID("sched1"), 32, false), &csCreate)
	assert.Equal(t, calSub{
		Name:            "Name 1",
		ScheduleID:      h.UUID("sched1"),
		ReminderMinutes: []int{32},
	}, csCreate.CreateUserCalendarSubscription.calSub)
	assert.NotEmpty(t, csCreate.CreateUserCalendarSubscription.ID)

	// update
	doQL(t, fmt.Sprintf(`
		mutation {
		  updateUserCalendarSubscription(input: {
			id: "%s"
			name: "%s"
			disabled: %v
		  })
		}
	`, csCreate.CreateUserCalendarSubscription.ID, "updated", true), nil)

	// find one
	var csFindOne struct {
		UserCalendarSubscription calSubWithID
	}
	doQL(t, fmt.Sprintf(`
		query {
			userCalendarSubscription(id: "%s") {
				id
				name
				disabled
			}
		}
	`, csCreate.CreateUserCalendarSubscription.ID), &csFindOne)
	assert.Equal(t, calSub{
		Name:     "updated",
		Disabled: true,
	}, csFindOne.UserCalendarSubscription.calSub)
	assert.Equal(t, csCreate.CreateUserCalendarSubscription.ID, csFindOne.UserCalendarSubscription.ID)

	var csFindMany struct {
		User struct {
			CalendarSubscriptions []calSub
		}
	}

	// find many
	doQL(t, `
		query{
			user {
				calendarSubscriptions {
				  id
				  name
				  reminderMinutes
				  scheduleID
				  disabled
				}
		    }
		}
	`, &csFindMany)
	assert.Equal(t, []calSub{{
		Name:            "updated",
		Disabled:        true,
		ReminderMinutes: []int{32},
		ScheduleID:      h.UUID("sched1"),
	}}, csFindMany.User.CalendarSubscriptions)

	// delete
	var delete struct {
		DeleteAll bool
	}
	doQL(t, fmt.Sprintf(`
		mutation {
			deleteAll(input: [{
				id: "%s"
				type: calendarSubscription
			}])
		}
	`, csCreate.CreateUserCalendarSubscription.ID), &delete)
	assert.True(t, delete.DeleteAll)

	// ensure delete happened
	doQL(t, `
		query{
			user {
				calendarSubscriptions {
				  id
				  name
				  reminderMinutes
				  scheduleID
				  disabled
				}
		    }
		}
	`, &csFindMany)
	assert.Equal(t, []calSub{}, csFindMany.User.CalendarSubscriptions)
}
