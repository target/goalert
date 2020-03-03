package smoketest

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

// TestGraphQLCreateSchedule tests that all steps for creating a schedule (without default rotation) are carried out without any errors.
func TestGraphQLCreateSchedule(t *testing.T) {
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
				ScheduleID string
				Target     struct{ ID string }
			}
		}
	}

	doQL(`mutation {
					createSchedule(
						input: {
							name: "defsslt"
							description: "default testing"
							timeZone: "America/Chicago"
						}
					) {
						id
						name
						targets {
							scheduleID
							target {
								id
							}
						}
					}
				}
	`, &sched)

	sID := sched.CreateSchedule.ID
	t.Log("Created Schedule ID :", sID)

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
						type
					}
				}
			}
		}
	
	`, sID), &newSched)

	t.Log("Number of rotations:", newSched)

	if len(newSched.Schedule.Targets) != 0 {
		t.Errorf("got %d rotations; want 0", len(newSched.Schedule.Targets))
	}
}
