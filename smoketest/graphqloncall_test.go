package smoketest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/smoketest/harness"
	"github.com/target/goalert/user"
)

// TestGraphQLOnCall tests the logic behind `User.is_on_call`.
func TestGraphQLOnCall(t *testing.T) {
	t.Parallel()

	h := harness.NewHarness(t, "", "escalation-policy-step-reorder")
	defer h.Close()

	doQL := func(t *testing.T, query string, res interface{}) {
		g := h.GraphQLQueryT(t, query, "/api/graphql")
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
			t.Fatal(err)
		}
	}

	var idCounter int

	check := func(name, tmplStr string, isUser1OnCall, isUser2OnCall bool) {
		t.Helper()
		var data struct {
			UniqName string
			User1    *user.User
			User2    *user.User
		}
		data.UniqName = fmt.Sprintf("generated%d", idCounter)
		idCounter++

		u1, u2 := h.CreateUser(), h.CreateUser()
		data.User1 = u1
		data.User2 = u2

		tmpl, err := template.New("mutation").Parse(tmplStr)
		require.NoError(t, err)

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, data)
		require.NoError(t, err)

		query := buf.String()

		t.Run(name, func(t *testing.T) {
			doQL(t, query, nil)
			h.Trigger()

			var onCall struct {
				User1, User2 struct {
					OnCallSteps []struct{ ID string }
				}
			}

			doQL(t, fmt.Sprintf(`
				query {
					user1: user(id: "%s") { onCallSteps{id} }
					user2: user(id: "%s") { onCallSteps{id} }
				}
			`, u1.ID, u2.ID), &onCall)

			assert.Equal(t, isUser1OnCall, len(onCall.User1.OnCallSteps) > 0, "User1 On-Call")
			assert.Equal(t, isUser2OnCall, len(onCall.User2.OnCallSteps) > 0, "User2 On-Call")
		})
	}

	// User directly on EP is always on call
	check("User EP Direct", `
		mutation{
			createService(input:{
				name: "{{.UniqName}}",
				newEscalationPolicy: {
					name: "{{.UniqName}}",
					steps: [{
						delayMinutes: 1,
						targets: [{type: user, id: "{{.User1.ID}}" }]
					}]
				}
			}){id}
		}
	`, true, false)

	// Active participant directly on EP is always on call
	check("User EP Rotation Direct", `
		mutation{
			createService(input:{
				name: "{{.UniqName}}",
				newEscalationPolicy: {
					name: "{{.UniqName}}",
					steps: [{
						delayMinutes: 1,
						newRotation: {
							name: "{{.UniqName}}",
							type: weekly,
							start: "2006-01-02T15:04:05Z",
							timeZone: "UTC",
							userIDs: ["{{.User1.ID}}"]
						}
					}]
				}
			}){id}
		}
	`, true, false)

	// EP -> Schedule, where there is an active ADD for a user
	check("User EP Schedule Add Override", `
		mutation{
			createService(input:{
				name: "{{.UniqName}}",
				newEscalationPolicy: {
					name: "{{.UniqName}}",
					steps: [{
						delayMinutes: 1,
						newSchedule: {
							name: "{{.UniqName}}",
							timeZone: "UTC",
							newUserOverrides: [{
								addUserID: "{{.User1.ID}}",
								start: "1006-01-02T15:04:05Z",
								end: "4006-01-02T15:04:05Z"
							}]
						}
					}]
				}
			}){id}
		}
	`, true, false)

	// // Active schedule rule, user is replaced
	// check("User EP Schedule Replace Override", `
	// 	escalation_policies: [{ id_placeholder: "ep", name: "generated", description: "1"}]
	// 	escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
	// 	services: [{id_placeholder: "svc", description: "ok", name: "generated", escalation_policy_id: "ep"}]
	// 	schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generated", description: "1"}]
	// 	schedule_rules: [{target:{target_type:user, target_id:"u2"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
	// 	user_overrides: [{add_user_id: "u1", remove_user_id: "u2", start_time: "1006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
	// `, true, false)

	// // Same scenario, user is NOT replaced (no override)
	// check("User EP Schedule Replace Override Absent", `
	// 		escalation_policies: [{ id_placeholder: "ep", name: "generated", description: "1"}]
	// 		escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
	// 		services: [{id_placeholder: "svc", description: "ok", name: "generated", escalation_policy_id: "ep"}]
	// 		schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generated", description: "1"}]
	// 		schedule_rules: [{target:{target_type:user, target_id:"u2"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
	// 	`, false, true)

	// // Active schedule rule, active rotation participant is replaced
	// check("User EP Schedule Replace Rotation Override", `
	// 	escalation_policies: [{ id_placeholder: "ep", name: "generated", description: "1"}]
	// 	escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
	// 	services: [{id_placeholder: "svc", description: "ok", name: "generated", escalation_policy_id: "ep"}]
	// 	schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generated", description: "1"}]
	// 	schedule_rules: [{target:{target_type:rotation, target_id:"rot"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
	// 	rotations: [{id_placeholder: "rot", time_zone: "UTC", shift_length: 1, type: weekly, start: "2006-01-02T15:04:05Z", name: "generated", description: "1"}]
	// 	rotation_participants: [{rotation_id: "rot", user_id: "u2"}]
	// 	user_overrides: [{add_user_id: "u1", remove_user_id: "u2", start_time: "1006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
	// `, true, false)

	// // Active schedule rule, active rotation participant is NOT replaced (no override)
	// check("User EP Schedule Replace Rotation Override Absent", `
	// 	escalation_policies: [{ id_placeholder: "ep", name: "generated", description: "1"}]
	// 	escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
	// 	services: [{id_placeholder: "svc", description: "ok", name: "generated", escalation_policy_id: "ep"}]
	// 	schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generated", description: "1"}]
	// 	schedule_rules: [{target:{target_type:rotation, target_id:"rot"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
	// 	rotations: [{id_placeholder: "rot", time_zone: "UTC", shift_length: 1, type: weekly, start: "2006-01-02T15:04:05Z", name: "generated", description: "1"}]
	// 	rotation_participants: [{rotation_id: "rot", user_id: "u2"}]
	// `, false, true)

	// // Active schedule rule, active rotation participant is removed
	// check("User EP Schedule Remove Rotation Override", `
	// 	escalation_policies: [{ id_placeholder: "ep", name: "generated", description: "1"}]
	// 	escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
	// 	services: [{id_placeholder: "svc", description: "ok", name: "generated", escalation_policy_id: "ep"}]
	// 	schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generated", description: "1"}]
	// 	schedule_rules: [{target:{target_type:rotation, target_id:"rot"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
	// 	rotations: [{id_placeholder: "rot", time_zone: "UTC", shift_length: 1, type: weekly, start: "2006-01-02T15:04:05Z", name: "generated", description: "1"}]
	// 	rotation_participants: [{rotation_id: "rot", user_id: "u2"}]
	// 	user_overrides: [{ remove_user_id: "u2", start_time: "1006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
	// `, false, false)

	// // Active schedule rule, user is removed
	// check("User EP Schedule Remove Override", `
	// 	escalation_policies: [{ id_placeholder: "ep", name: "generated", description: "1"}]
	// 	escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
	// 	services: [{id_placeholder: "svc", description: "ok", name: "generated", escalation_policy_id: "ep"}]
	// 	schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generated", description: "1"}]
	// 	schedule_rules: [{target:{target_type:user, target_id:"u2"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
	// 	user_overrides: [{remove_user_id: "u2", start_time: "1006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
	// `, false, false)

}
