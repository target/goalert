package smoketest

import (
	"encoding/json"
	"fmt"
	"github.com/target/goalert/smoketest/harness"
	"testing"
	"time"
)

// TestGraphQLUpdateRotation tests that all steps like creating and updating rotations are carried out without any errors.
func TestGraphQLUpdateRotation(t *testing.T) {
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
				name: "default",
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
	var rotation struct {
		CreateOrUpdateRotation struct {
			Rotation struct {
				ID   string
				Name string
			}
		}
	}
	doQL(fmt.Sprintf(`
		mutation {
			createOrUpdateRotation(input:{
				id: "%s",
				name: "default",
				start: "2017-08-15T19:00:00Z",
				type: daily,
				shift_length: 2,
				schedule_id: "%s"
			}){
				rotation {
					id
					name
				}
			}
		}
	
	`, sched.CreateSchedule.Rotations[0].ID, sID), &rotation)

	var newSched struct {
		Schedule struct {
			Rotations []struct {
				ShiftLength int `json:"shift_length"`
			}
		}
	}
	doQL(fmt.Sprintf(`
		query {
			schedule(id: "%s") {
				rotations {
					id
					shift_length
				}
			}
		}
	
	`, sID), &newSched)

	if len(newSched.Schedule.Rotations) != 1 {
		t.Errorf("got %d rotations; want 1", len(newSched.Schedule.Rotations))
	}
	if newSched.Schedule.Rotations[0].ShiftLength != 2 {
		t.Errorf("got shift_length of %d; want 2", newSched.Schedule.Rotations[0].ShiftLength)
	}
}
