package smoketest

import (
	"encoding/json"
	"fmt"
	"github.com/target/goalert/smoketest/harness"
	"testing"
	"time"
)

// TestGraphQLCreateScheduleWithDefaultRotation tests that all steps for creating a schedule with default rotation are carried out without any errors.
func TestGraphQLCreateScheduleWithDefaultRotation(t *testing.T) {
	t.Parallel()

	const sql = `
	insert into users (id, name, email)
	values
		({{uuid "u1"}}, 'bob', 'joe'),
		({{uuid "u2"}}, 'ben', 'josh');
`
	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	doQL := func(query string, res interface{}) {
		g := h.GraphQLQuery(query)
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

	var sched struct {
		CreateSchedule struct {
			ID        string
			Rotations []struct{ ID string }
		}
	}

	doQL(fmt.Sprintf(`
		mutation {
			createSchedule(input:{
				name: "default_testing",
				description: "default testing",
				time_zone: "America/Chicago",
				default_rotation: {
					type: daily,
					start_time: "%s",
    				shift_length:1,
  				}
			}){
				id
				rotations {
					id
				}
			}
		}
	
	`, time.Now().Format(time.RFC3339)), &sched)

	sID := sched.CreateSchedule.ID
	t.Log("Created Schedule ID :", sID)

	var newSched struct {
		Schedule struct {
			Rotations []struct{}
		}
	}
	doQL(fmt.Sprintf(`
		query {
			schedule(id: "%s") {
				rotations {
					id
				}
			}
		}
	
	`, sID), &newSched)

	t.Log("Number of rotations:", newSched)

	if len(newSched.Schedule.Rotations) != 1 {
		t.Errorf("got %d rotations; want 1", len(newSched.Schedule.Rotations))
	}
}
