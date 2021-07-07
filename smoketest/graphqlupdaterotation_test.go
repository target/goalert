package smoketest

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/smoketest/harness"
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

	u1UUID := h.UUID("u1")
	u2UUID := h.UUID("u2")

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

	var sched struct {
		CreateSchedule struct {
			ID      string
			Name    string
			Targets []struct {
				Target struct {
					ID string
				}
			}
		}
	}
	doQL(fmt.Sprintf(`
		mutation {
			createSchedule(
				input: {
					name: "default"
					description: "default testing"
					timeZone: "America/Chicago"
					targets: {
						newRotation: {
							name: "old name"
							description: "old description"
							timeZone: "America/Chicago"
							start: "2020-02-04T12:08:25-06:00"
							type: daily
							shiftLength: 6
							userIDs: ["%s"]
						}
						rules: {
							start: "00:00"
							end: "23:00"
							weekdayFilter: [true, true, true, true, true, true, true]
						}
					}
				}
			) {
				id
				name
				targets {
					target {
						id
					}
				}
			}
		}
	`, u1UUID), &sched)

	rotationID := sched.CreateSchedule.Targets[0].Target.ID

	doQL(fmt.Sprintf(`
		mutation {
			updateRotation(input:{
				id: "%s",
				name: "new name",
				description: "new description"
				timeZone: "America/New_York"
				start: "1997-11-26T12:08:25-05:00"
				type: hourly
				shiftLength: 12
				activeUserIndex: 0
				userIDs: ["%s", "%s"]
			})
		}
	
	`, rotationID, u1UUID, u2UUID), nil)

	var newSched struct {
		Schedule struct {
			ID      string
			Name    string
			Targets []struct {
				ScheduleID string
				Target     struct{ ID string }
			}
		}
	}
	doQL(fmt.Sprintf(`
		query {
			schedule(id: "%s") {
				targets {
					target {
						id
					}
				}
			}
		}
	
	`, sched.CreateSchedule.ID), &newSched)

	if len(newSched.Schedule.Targets) != 1 {
		t.Errorf("got %d rotations; want 1", len(newSched.Schedule.Targets))
	}

	var updatedRotation struct {
		Rotation struct {
			Name            string
			Description     string
			TimeZone        string
			Start           string
			Type            string
			ShiftLength     int
			ActiveUserIndex int
			Users           []struct {
				ID string
			}
		}
	}
	doQL(fmt.Sprintf(`
		query{
		rotation(id: "%s"){
			name
			description
			timeZone
			start
			type
			shiftLength
			activeUserIndex
			users {
				id
			}
		}
	}`, rotationID), &updatedRotation)

	assert.Equal(t, "new name", updatedRotation.Rotation.Name)
	assert.Equal(t, "new description", updatedRotation.Rotation.Description)
	assert.Equal(t, "America/New_York", updatedRotation.Rotation.TimeZone)
	assert.Equal(t, "1997-11-26T17:08:00Z", updatedRotation.Rotation.Start) // truncate to minute
	assert.Equal(t, "hourly", updatedRotation.Rotation.Type)
	assert.Equal(t, 12, updatedRotation.Rotation.ShiftLength)
	assert.Equal(t, 0, updatedRotation.Rotation.ActiveUserIndex)
	assert.Equal(t, u1UUID, updatedRotation.Rotation.Users[0].ID)
	assert.Equal(t, u2UUID, updatedRotation.Rotation.Users[1].ID)
}
