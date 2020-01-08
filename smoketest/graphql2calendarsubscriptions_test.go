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

		insert into schedules (id, name, time_zone) 
		values ({{uuid "sched1"}}, 'default', 'America/Chicago');

		insert into user_calendar_subscriptions (id, name, user_id, config, schedule_id)
		values ({{uuid "cs1"}}, 'test1', {{uuid "user"}}, '{ "notification_minutes": [2, 4, 8, 16]}', {{uuid "sched1"}});

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

	var cs struct {
		UserCalendarSubscription struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
	}

	var csCreate struct {
		UserCreateCalendarSubscription struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			ScheduleID string `json:"scheduleID"`
		}
	}

	var csMany struct {
		User struct {
			CalendarSubscriptions []struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				Disabled   bool   `json:"disabled"`
				ScheduleID string `json:"scheduleID"`
			}
		}
	}

	var delete struct {
		DeleteAll bool
	}

	// create
	doQL(t, fmt.Sprintf(`
		mutation {
		  userCreateCalendarSubscription(input: {
			name: "%s"
			scheduleID: "%s"
			notificationMinutes: [%d]
			disabled: %v
		  }) {
			ID
			name
			scheduleID
		  }
		}
	`, "Name 1", h.UUID("sched1"), 32, false), nil)

	// create
	doQL(t, fmt.Sprintf(`
		mutation {
		  userCreateCalendarSubscription(input: {
			name: "%s"
			scheduleID: "%s"
			notificationMinutes: [%d]
			disabled: %v
		  }) {
			ID
			name
			scheduleID
		  }
		}
	`, "Name 1", h.UUID("sched1"), 32, false), &csCreate)

	if csCreate.UserCreateCalendarSubscription.Name != "Name 1" {
		t.Fatalf("ERROR: CalendarSubscription %s Name=%s; want 'Name 1'", csCreate.UserCreateCalendarSubscription.ID, csCreate.UserCreateCalendarSubscription.Name)
	}
	if csCreate.UserCreateCalendarSubscription.ScheduleID != h.UUID("sched1") {
		t.Fatalf("ERROR: CalendarSubscription %s ScheduleID=%s; want %s", csCreate.UserCreateCalendarSubscription.ID, csCreate.UserCreateCalendarSubscription.ScheduleID, h.UUID("sched1"))
	}

	// update
	doQL(t, fmt.Sprintf(`
		mutation {
		  userUpdateCalendarSubscription(input: {
			id: "%s"
			name: "%s"
			disabled: %v
		  })
		}
	`, csCreate.UserCreateCalendarSubscription.ID, "updated", true), nil)

	// find one
	doQL(t, fmt.Sprintf(`
		query {
			userCalendarSubscription(id: "%s") {
				ID
				name
			}
		}
	`, csCreate.UserCreateCalendarSubscription.ID), &cs)

	if cs.UserCalendarSubscription.Name != "updated" {
		t.Fatalf("ERROR: CalendarSubscription %s Name=%s; want 'updated'", cs.UserCalendarSubscription.ID, cs.UserCalendarSubscription.Name)
	}

	// find many
	doQL(t, fmt.Sprintf(`
		query{
			user(id: "%s") {
				calendarSubscriptions {
				  ID
				  name
				  notificationMinutes
				  schedule {
					id
				  }
				  scheduleID
				}
		  }
		}
	`, h.UUID("user")), &csMany)

	if len(csMany.User.CalendarSubscriptions) != 2 {
		t.Fatalf("ERROR: Did not find all of the subscriptions created")
	}

	// delete
	doQL(t, fmt.Sprintf(`
		mutation {
			deleteAll(input: [{
				id: "%s"
				type: calendarSubscription
			}])
		}
	`, cs.UserCalendarSubscription.ID), &delete)

	if !delete.DeleteAll {
		t.Fatalf("ERROR: Did not delete CalendarSubscription %s", cs.UserCalendarSubscription.ID)
	}
}
