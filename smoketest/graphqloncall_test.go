package smoketest

import (
	"encoding/json"
	"fmt"
	"github.com/target/goalert/smoketest/harness"
	"strings"
	"testing"
)

// TestGraphQLOnCall tests the logic behind `User.is_on_call`.
func TestGraphQLOnCall(t *testing.T) {
	t.Parallel()

	h := harness.NewHarness(t, "", "escalation-policy-step-reorder")
	defer h.Close()

	doQL := func(t *testing.T, query string, res interface{}) {
		g := h.GraphQLQueryT(t, query, "/v1/graphql")
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

	check := func(name, input string, user1OnCall, user2OnCall bool) {
		u1 := h.CreateUser()
		u2 := h.CreateUser()
		input = strings.Replace(input, "u1", u1.ID, -1)
		input = strings.Replace(input, "u2", u2.ID, -1)
		input = strings.Replace(input, "generated", fmt.Sprintf("generated%d", idCounter), -1)
		idCounter++
		query := fmt.Sprintf(`
			mutation {
				createAll(input:{
					%s
				}) {
					services {id}
					escalation_policies {id}
					rotations {id}
					user_overrides {id}
					schedules {id}
				}
			}
			`, input)
		t.Run(name, func(t *testing.T) {

			var resp struct {
				CreateAll map[string][]struct{ ID string }
			}
			doQL(t, query, &resp)
			h.Trigger()

			var onCall struct {
				User struct {
					IsOnCall bool `json:"on_call"`
				}
			}

			doQL(t, fmt.Sprintf(`
				query {
					user(id: "%s") { on_call }
				
				}
			`, u1.ID), &onCall)

			if user1OnCall != onCall.User.IsOnCall {
				t.Fatalf("ERROR: User1 On-Call=%t; want %t", onCall.User.IsOnCall, user1OnCall)
			}

			doQL(t, fmt.Sprintf(`
				query {
					user(id: "%s") { on_call }
				
				}
			`, u2.ID), &onCall)

			if user2OnCall != onCall.User.IsOnCall {
				t.Fatalf("ERROR: User2 On-Call=%t; want %t", onCall.User.IsOnCall, user2OnCall)
			}

		})
	}

	// Randomly generate names, instead of hard-coding
	// User directly on EP is always on call
	check("User EP Direct", `
		escalation_policies: [{ id_placeholder: "ep", name: "generated", description: "1"}]
		escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: user, target_id: "u1" }] }]
		services: [{id_placeholder: "svc", description: "ok", name: "generated", escalation_policy_id: "ep"}]
	`, true, false)

	// Active participant directly on EP is always on call
	check("User EP Rotation Direct", `
		escalation_policies: [{ id_placeholder: "ep", name: "generated", description: "1"}]
		escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: rotation, target_id: "rot" }] }]
		services: [{id_placeholder: "svc", description: "ok", name: "generated", escalation_policy_id: "ep"}]
		rotations: [{id_placeholder: "rot", time_zone: "UTC", shift_length: 1, type: weekly, start: "2006-01-02T15:04:05Z", name: "generated", description: "1"}]
		rotation_participants: [{rotation_id: "rot", user_id: "u1"}]
	`, true, false)

	// EP -> Schedule, where there is an active ADD for a user
	check("User EP Schedule Add Override", `
		escalation_policies: [{ id_placeholder: "ep", name: "generated", description: "1"}]
		escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
		services: [{id_placeholder: "svc", description: "ok", name: "generated", escalation_policy_id: "ep"}]
		schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generated", description: "1"}]
		user_overrides: [{add_user_id: "u1", start_time: "1006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
	`, true, false)

	// Active schedule rule, user is replaced
	check("User EP Schedule Replace Override", `
		escalation_policies: [{ id_placeholder: "ep", name: "generated", description: "1"}]
		escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
		services: [{id_placeholder: "svc", description: "ok", name: "generated", escalation_policy_id: "ep"}]
		schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generated", description: "1"}]
		schedule_rules: [{target:{target_type:user, target_id:"u2"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
		user_overrides: [{add_user_id: "u1", remove_user_id: "u2", start_time: "1006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
	`, true, false)

	// Same scenario, user is NOT replaced (no override)
	check("User EP Schedule Replace Override Absent", `
			escalation_policies: [{ id_placeholder: "ep", name: "generated", description: "1"}]
			escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
			services: [{id_placeholder: "svc", description: "ok", name: "generated", escalation_policy_id: "ep"}]
			schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generated", description: "1"}]
			schedule_rules: [{target:{target_type:user, target_id:"u2"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
		`, false, true)

	// Active schedule rule, active rotation participant is replaced
	check("User EP Schedule Replace Rotation Override", `
		escalation_policies: [{ id_placeholder: "ep", name: "generated", description: "1"}]
		escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
		services: [{id_placeholder: "svc", description: "ok", name: "generated", escalation_policy_id: "ep"}]
		schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generated", description: "1"}]
		schedule_rules: [{target:{target_type:rotation, target_id:"rot"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
		rotations: [{id_placeholder: "rot", time_zone: "UTC", shift_length: 1, type: weekly, start: "2006-01-02T15:04:05Z", name: "generated", description: "1"}]
		rotation_participants: [{rotation_id: "rot", user_id: "u2"}]
		user_overrides: [{add_user_id: "u1", remove_user_id: "u2", start_time: "1006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
	`, true, false)

	// Active schedule rule, active rotation participant is NOT replaced (no override)
	check("User EP Schedule Replace Rotation Override Absent", `
		escalation_policies: [{ id_placeholder: "ep", name: "generated", description: "1"}]
		escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
		services: [{id_placeholder: "svc", description: "ok", name: "generated", escalation_policy_id: "ep"}]
		schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generated", description: "1"}]
		schedule_rules: [{target:{target_type:rotation, target_id:"rot"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
		rotations: [{id_placeholder: "rot", time_zone: "UTC", shift_length: 1, type: weekly, start: "2006-01-02T15:04:05Z", name: "generated", description: "1"}]
		rotation_participants: [{rotation_id: "rot", user_id: "u2"}]
	`, false, true)

	// Active schedule rule, active rotation participant is removed
	check("User EP Schedule Remove Rotation Override", `
		escalation_policies: [{ id_placeholder: "ep", name: "generated", description: "1"}]
		escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
		services: [{id_placeholder: "svc", description: "ok", name: "generated", escalation_policy_id: "ep"}]
		schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generated", description: "1"}]
		schedule_rules: [{target:{target_type:rotation, target_id:"rot"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
		rotations: [{id_placeholder: "rot", time_zone: "UTC", shift_length: 1, type: weekly, start: "2006-01-02T15:04:05Z", name: "generated", description: "1"}]
		rotation_participants: [{rotation_id: "rot", user_id: "u2"}]
		user_overrides: [{ remove_user_id: "u2", start_time: "1006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
	`, false, false)

	// Active schedule rule, user is removed
	check("User EP Schedule Remove Override", `
		escalation_policies: [{ id_placeholder: "ep", name: "generated", description: "1"}]
		escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
		services: [{id_placeholder: "svc", description: "ok", name: "generated", escalation_policy_id: "ep"}]
		schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generated", description: "1"}]
		schedule_rules: [{target:{target_type:user, target_id:"u2"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
		user_overrides: [{remove_user_id: "u2", start_time: "1006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
	`, false, false)

}
