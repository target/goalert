package smoketest

import (
	"encoding/json"
	"fmt"
	"github.com/target/goalert/engine/resolver"
	"github.com/target/goalert/smoketest/harness"
	"strings"
	"testing"
)

// TestGraphQLOnCallAssignments tests the logic behind `User.is_on_call`.
func TestGraphQLOnCallAssignments(t *testing.T) {
	t.Parallel()

	h := harness.NewHarness(t, "", "escalation-policy-step-reorder")
	defer h.Close()

	doQL := func(t *testing.T, silent bool, query string, res interface{}) {
		g := h.GraphQLQueryT(t, query, "/v1/graphql")
		for _, err := range g.Errors {
			t.Error("GraphQL Error:", err.Message)
		}
		if len(g.Errors) > 0 {
			t.Fatal("errors returned from GraphQL")
		}
		if !silent {
			t.Log("Response:", string(g.Data))
		}

		if res == nil {
			return
		}
		err := json.Unmarshal(g.Data, &res)
		if err != nil {
			t.Fatal(err)
		}
	}

	type asnID struct {
		Svc, EP, Rot, Sched string
		Step                int
	}

	getID := func(a resolver.OnCallAssignment) asnID {
		return asnID{
			Svc:   a.ServiceName,
			EP:    a.EPName,
			Rot:   a.RotationName,
			Sched: a.ScheduleName,
			Step:  a.Level,
		}
	}

	var idCounter int
	check := func(name, input string, user1OnCall, user2OnCall []resolver.OnCallAssignment) {
		u1 := h.CreateUser()
		u2 := h.CreateUser()
		rep := strings.NewReplacer(
			"generatedA", fmt.Sprintf("generatedA%d", idCounter),
			"generatedB", fmt.Sprintf("generatedB%d", idCounter),
		)
		idCounter++

		for i, oc := range user1OnCall {
			oc.EPName = rep.Replace(oc.EPName)
			oc.RotationName = rep.Replace(oc.RotationName)
			oc.ScheduleName = rep.Replace(oc.ScheduleName)
			oc.ServiceName = rep.Replace(oc.ServiceName)
			user1OnCall[i] = oc
		}

		for i, oc := range user2OnCall {
			oc.EPName = rep.Replace(oc.EPName)
			oc.RotationName = rep.Replace(oc.RotationName)
			oc.ScheduleName = rep.Replace(oc.ScheduleName)
			oc.ServiceName = rep.Replace(oc.ServiceName)
			user2OnCall[i] = oc
		}

		input = strings.Replace(input, "u1", u1.ID, -1)
		input = strings.Replace(input, "u2", u2.ID, -1)
		input = rep.Replace(input)
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
			doQL(t, false, query, &resp)
			h.Trigger()

			var onCall struct {
				User struct {
					OnCallAssignments []resolver.OnCallAssignment `json:"on_call_assignments"`
				}
			}

			var hasFailure bool

			checkUser := func(name, uid string) {

				t.Run("User_"+name, func(t *testing.T) {
					doQL(t, false, fmt.Sprintf(`
						query {
							user(id: "%s") { on_call_assignments{
								escalation_policy_name
								escalation_policy_step_number
								is_active
								rotation_name
								schedule_name
								service_name
								user_id
							} }
						
						}
					`, uid), &onCall)

					m := make(map[asnID]resolver.OnCallAssignment, len(onCall.User.OnCallAssignments))
					checked := make(map[asnID]bool)
					for _, a := range onCall.User.OnCallAssignments {
						m[getID(a)] = a
					}
					var asn []resolver.OnCallAssignment
					switch name {
					case "u1":
						asn = user1OnCall
					case "u2":
						asn = user2OnCall
					}

					for _, a := range asn {
						id := getID(a)
						checked[id] = true
						resp, ok := m[id]
						if !ok {
							hasFailure = true
							t.Errorf("got nil, want assignment %+v", id)
							continue
						}

						if resp.UserID != uid {
							hasFailure = true
							t.Errorf("Bad UserID for %+v: got %s; want %s", id, resp.UserID, uid)
						}

						if resp.IsActive != a.IsActive {
							hasFailure = true
							t.Errorf("Wrong active state for %+v: got %t; want %t", id, resp.IsActive, a.IsActive)
						}
					}
					for aID := range m {
						if checked[aID] {
							continue
						}
						hasFailure = true
						t.Errorf("got unexpected assignment: %+v", aID)
					}
				})
			}

			checkUser("u1", u1.ID)
			checkUser("u2", u2.ID)

			if hasFailure {
				t.Fatal()
			}

		})
	}

	// User directly on EP is always on call
	check("User EP Direct", `
			escalation_policies: [{ id_placeholder: "ep", name: "generatedA", description: "1"}]
			escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: user, target_id: "u1" }] }]
			services: [{id_placeholder: "svc", description: "ok", name: "generatedA", escalation_policy_id: "ep"}]
		`,
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", Level: 0, IsActive: true},
		},
		nil,
	)

	// Active participant directly on EP is always on call
	check("User EP Rotation Direct", `
			escalation_policies: [{ id_placeholder: "ep", name: "generatedA", description: "1"}]
			escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: rotation, target_id: "rot" }] }]
			services: [{id_placeholder: "svc", description: "ok", name: "generatedA", escalation_policy_id: "ep"}]
			rotations: [{id_placeholder: "rot", time_zone: "UTC", shift_length: 1, type: weekly, start: "2006-01-02T15:04:05Z", name: "generatedA", description: "1"}]
			rotation_participants: [{rotation_id: "rot", user_id: "u1"}]
		`,
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", RotationName: "generatedA", Level: 0, IsActive: true},
		},
		nil,
	)

	// Active participant directly on EP is always on call, rotation directly on EP but no participant, user has no assignments
	check("Only One User EP Rotation Direct", `
			escalation_policies: [{ id_placeholder: "ep", name: "generatedA", description: "1"}]
			escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: rotation, target_id: "rot" }, {target_type: rotation, target_id: "rot2" }] }]
			services: [{id_placeholder: "svc", description: "ok", name: "generatedA", escalation_policy_id: "ep"}]
			rotations: [{id_placeholder: "rot", time_zone: "UTC", shift_length: 1, type: weekly, start: "3006-01-02T15:04:05Z", name: "generatedA", description: "1"},
						{id_placeholder: "rot2", time_zone: "UTC", shift_length: 1, type: weekly, start: "2016-01-02T15:04:05Z", name: "generatedB", description: "2"} ]
			rotation_participants: [{rotation_id: "rot2", user_id: "u2"}]
		`,
		nil,
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", RotationName: "generatedB", Level: 0, IsActive: true},
		},
	)

	// Different users on different rotations, users are on call but with different assignment rotations
	check("Multiple Users EP Rotation Direct", `
			escalation_policies: [{ id_placeholder: "ep", name: "generatedA", description: "1"}]
			escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: rotation, target_id: "rot" }, {target_type: rotation, target_id: "rot2" }] }]
			services: [{id_placeholder: "svc", description: "ok", name: "generatedA", escalation_policy_id: "ep"}]
			rotations: [{id_placeholder: "rot", time_zone: "UTC", shift_length: 1, type: weekly, start: "2006-01-02T15:04:05Z", name: "generatedA", description: "1"},
						{id_placeholder: "rot2", time_zone: "UTC", shift_length: 1, type: weekly, start: "2016-01-02T15:04:05Z", name: "generatedB", description: "2"} ]
			rotation_participants: [{rotation_id: "rot", user_id: "u1"}, {rotation_id: "rot2", user_id: "u2"}]
		`,
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", RotationName: "generatedA", Level: 0, IsActive: true},
		},
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", RotationName: "generatedB", Level: 0, IsActive: true},
		},
	)

	// EP -> Schedule, where there is an active ADD for a user
	check("User EP Schedule Add Override", `
			escalation_policies: [{ id_placeholder: "ep", name: "generatedA", description: "1"}]
			escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
			services: [{id_placeholder: "svc", description: "ok", name: "generatedA", escalation_policy_id: "ep"}]
			schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generatedA", description: "1"}]
			user_overrides: [{add_user_id: "u1", start_time: "1006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
		`,
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: true},
		},
		nil,
	)

	// EP -> Schedule, where there is an inactive ADD for a user
	check("User EP Schedule Inactive Add Override", `
			escalation_policies: [{ id_placeholder: "ep", name: "generatedA", description: "1"}]
			escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
			services: [{id_placeholder: "svc", description: "ok", name: "generatedA", escalation_policy_id: "ep"}]
			schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generatedA", description: "1"}]
			user_overrides: [{add_user_id: "u1", start_time: "3006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
		`,
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: false},
		},
		nil,
	)

	// Active schedule rule, user is replaced
	check("User EP Schedule Replace Override", `
			escalation_policies: [{ id_placeholder: "ep", name: "generatedA", description: "1"}]
			escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
			services: [{id_placeholder: "svc", description: "ok", name: "generatedA", escalation_policy_id: "ep"}]
			schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generatedA", description: "1"}]
			schedule_rules: [{target:{target_type:user, target_id:"u2"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
			user_overrides: [{add_user_id: "u1", remove_user_id: "u2", start_time: "1006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
		`,
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: true},
		},
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: false},
		},
	)

	// Active schedule rule, user is replaced but in the future (inactive replacement)
	check("User EP Schedule Inactive Replace Override", `
			escalation_policies: [{ id_placeholder: "ep", name: "generatedA", description: "1"}]
			escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
			services: [{id_placeholder: "svc", description: "ok", name: "generatedA", escalation_policy_id: "ep"}]
			schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generatedA", description: "1"}]
			schedule_rules: [{target:{target_type:user, target_id:"u2"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
			user_overrides: [{add_user_id: "u1", remove_user_id: "u2", start_time: "3006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
		`,
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: false},
		},
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: true},
		},
	)

	// Same scenario, user is NOT replaced (no override)
	check("User EP Schedule Replace Override Absent", `
			escalation_policies: [{ id_placeholder: "ep", name: "generatedA", description: "1"}]
			escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
			services: [{id_placeholder: "svc", description: "ok", name: "generatedA", escalation_policy_id: "ep"}]
			schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generatedA", description: "1"}]
			schedule_rules: [{target:{target_type:user, target_id:"u2"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
		`,
		[]resolver.OnCallAssignment{},
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: true},
		},
	)

	// Same scenario, user is NOT replaced (no override), inactive schedule rule
	check("User EP Schedule No Days Replace Override Absent", `
		escalation_policies: [{ id_placeholder: "ep", name: "generatedA", description: "1"}]
		escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
		services: [{id_placeholder: "svc", description: "ok", name: "generatedA", escalation_policy_id: "ep"}]
		schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generatedA", description: "1"}]
		schedule_rules: [{target:{target_type:user, target_id:"u2"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: false, monday:false, tuesday:false, wednesday: false, thursday: false, friday: false, saturday: false}]
	`,
		[]resolver.OnCallAssignment{},
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: false},
		},
	)

	// Active schedule rule, active rotation participant is replaced
	check("User EP Schedule Replace Rotation Override", `
			escalation_policies: [{ id_placeholder: "ep", name: "generatedA", description: "1"}]
			escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
			services: [{id_placeholder: "svc", description: "ok", name: "generatedA", escalation_policy_id: "ep"}]
			schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generatedA", description: "1"}]
			schedule_rules: [{target:{target_type:rotation, target_id:"rot"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
			rotations: [{id_placeholder: "rot", time_zone: "UTC", shift_length: 1, type: weekly, start: "2006-01-02T15:04:05Z", name: "generatedA", description: "1"}]
			rotation_participants: [{rotation_id: "rot", user_id: "u2"}]
			user_overrides: [{add_user_id: "u1", remove_user_id: "u2", start_time: "1006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
		`,
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: true},
		},
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: false},
		},
	)

	// Active schedule rule, active rotation participant is replaced
	check("User EP Schedule Replace Rotation Override", `
	escalation_policies: [{ id_placeholder: "ep", name: "generatedA", description: "1"}]
	escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
	services: [{id_placeholder: "svc", description: "ok", name: "generatedA", escalation_policy_id: "ep"}]
	schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generatedA", description: "1"}]
	schedule_rules: [{target:{target_type:rotation, target_id:"rot"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
	rotations: [{id_placeholder: "rot", time_zone: "UTC", shift_length: 1, type: weekly, start: "2006-01-02T15:04:05Z", name: "generatedA", description: "1"}]
	rotation_participants: [{rotation_id: "rot", user_id: "u2"}]
	user_overrides: [{add_user_id: "u1", remove_user_id: "u2", start_time: "1006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
`,
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: true},
		},
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: false},
		},
	)

	// Active schedule rule, active rotation participant is replaced with an inactive replace override
	check("User EP Schedule Replace Rotation Override (Inactive)", `
		escalation_policies: [{ id_placeholder: "ep", name: "generatedA", description: "1"}]
		escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
		services: [{id_placeholder: "svc", description: "ok", name: "generatedA", escalation_policy_id: "ep"}]
		schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generatedA", description: "1"}]
		schedule_rules: [{target:{target_type:rotation, target_id:"rot"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
		rotations: [{id_placeholder: "rot", time_zone: "UTC", shift_length: 1, type: weekly, start: "2006-01-02T15:04:05Z", name: "generatedA", description: "1"}]
		rotation_participants: [{rotation_id: "rot", user_id: "u2"}]
		user_overrides: [{add_user_id: "u1", remove_user_id: "u2", start_time: "3006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
	`,
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: false},
		},
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: true},
		},
	)

	// Same as above, but no service assignment
	check("User EP Schedule Replace Rotation Override No Service", `
			escalation_policies: [{ id_placeholder: "ep", name: "generatedA", description: "1"}]
			escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
			schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generatedA", description: "1"}]
			schedule_rules: [{target:{target_type:rotation, target_id:"rot"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
			rotations: [{id_placeholder: "rot", time_zone: "UTC", shift_length: 1, type: weekly, start: "2006-01-02T15:04:05Z", name: "generatedA", description: "1"}]
			rotation_participants: [{rotation_id: "rot", user_id: "u2"}]
			user_overrides: [{add_user_id: "u1", remove_user_id: "u2", start_time: "1006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
		`,
		[]resolver.OnCallAssignment{},
		[]resolver.OnCallAssignment{},
	)

	// Same as above, but 2 service assignments
	check("User EP Schedule Replace Rotation Override Double Service", `
			escalation_policies: [{ id_placeholder: "ep", name: "generatedA", description: "1"}]
			escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
			services: [{id_placeholder: "svc", description: "ok", name: "generatedA", escalation_policy_id: "ep"},{description: "ok", name: "generatedB", escalation_policy_id: "ep"}]
			schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generatedA", description: "1"}]
			schedule_rules: [{target:{target_type:rotation, target_id:"rot"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
			rotations: [{id_placeholder: "rot", time_zone: "UTC", shift_length: 1, type: weekly, start: "2006-01-02T15:04:05Z", name: "generatedA", description: "1"}]
			rotation_participants: [{rotation_id: "rot", user_id: "u2"}]
			user_overrides: [{add_user_id: "u1", remove_user_id: "u2", start_time: "1006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
		`,
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: true},
			{ServiceName: "generatedB", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: true},
		},
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: false},
			{ServiceName: "generatedB", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: false},
		},
	)

	// Active schedule rule, active rotation participant is NOT replaced (no override)
	check("User EP Schedule Replace Rotation Override Absent", `
			escalation_policies: [{ id_placeholder: "ep", name: "generatedA", description: "1"}]
			escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
			services: [{id_placeholder: "svc", description: "ok", name: "generatedA", escalation_policy_id: "ep"}]
			schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generatedA", description: "1"}]
			schedule_rules: [{target:{target_type:rotation, target_id:"rot"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
			rotations: [{id_placeholder: "rot", time_zone: "UTC", shift_length: 1, type: weekly, start: "2006-01-02T15:04:05Z", name: "generatedA", description: "1"}]
			rotation_participants: [{rotation_id: "rot", user_id: "u2"}]
		`,
		[]resolver.OnCallAssignment{},
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: true},
		},
	)

	// Active schedule rule, active rotation participant is removed
	check("User EP Schedule Remove Rotation Override", `
			escalation_policies: [{ id_placeholder: "ep", name: "generatedA", description: "1"}]
			escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
			services: [{id_placeholder: "svc", description: "ok", name: "generatedA", escalation_policy_id: "ep"}]
			schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generatedA", description: "1"}]
			schedule_rules: [{target:{target_type:rotation, target_id:"rot"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
			rotations: [{id_placeholder: "rot", time_zone: "UTC", shift_length: 1, type: weekly, start: "2006-01-02T15:04:05Z", name: "generatedA", description: "1"}]
			rotation_participants: [{rotation_id: "rot", user_id: "u2"}]
			user_overrides: [{ remove_user_id: "u2", start_time: "1006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
		`,
		[]resolver.OnCallAssignment{},
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: false},
		},
	)

	// Active schedule rule, user is removed
	check("User EP Schedule Remove Override", `
			escalation_policies: [{ id_placeholder: "ep", name: "generatedA", description: "1"}]
			escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
			services: [{id_placeholder: "svc", description: "ok", name: "generatedA", escalation_policy_id: "ep"}]
			schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generatedA", description: "1"}]
			schedule_rules: [{target:{target_type:user, target_id:"u2"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
			user_overrides: [{remove_user_id: "u2", start_time: "1006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
		`,
		[]resolver.OnCallAssignment{},
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: false},
		},
	)

	// Multiple add overrides, active schedule rules
	check("User EP Schedule Multiple Overrides", `
			escalation_policies: [{ id_placeholder: "ep", name: "generatedA", description: "1"}]
			escalation_policy_steps: [{escalation_policy_id: "ep", delay_minutes: 1, targets: [{target_type: schedule, target_id: "s" }] }]
			services: [{id_placeholder: "svc", description: "ok", name: "generatedA", escalation_policy_id: "ep"}]
			schedules: [{id_placeholder: "s", time_zone: "UTC", name: "generatedA", description: "1"}]
			schedule_rules: [{target:{target_type:user, target_id:"u2"}, start:"00:00", end:"23:59", schedule_id: "s", sunday: true, monday:true, tuesday:true, wednesday: true, thursday: true, friday: true, saturday: true}]
			user_overrides: [{add_user_id: "u1", start_time: "1006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"},
							 {add_user_id: "u2", start_time: "1006-01-02T15:04:05Z", end_time: "4006-01-02T15:04:05Z", target_type: schedule, target_id: "s"}]
		`,
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: true},
		},
		[]resolver.OnCallAssignment{
			{ServiceName: "generatedA", EPName: "generatedA", ScheduleName: "generatedA", Level: 0, IsActive: true},
		},
	)

}
