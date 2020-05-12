package smoketest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/smoketest/harness"
	"github.com/target/goalert/user"
)

var onCallAsnQueryTmpl = template.Must(template.New("query").Parse(`
	query {
		{{- range $id, $usr := .}}
			{{$id}}: user(id: "{{$usr.ID}}") { onCallSteps{stepNumber, escalationPolicy {name, assignedTo { name }}} }
		{{- end}}
	}
`))

// TestGraphQLOnCallAssignments tests the logic behind `User.is_on_call`.
func TestGraphQLOnCallAssignments(t *testing.T) {
	t.Parallel()

	sql := `insert into escalation_policies (id, name) 
					values ({{uuid "eid"}}, 'esc policy');`

	type onCallAssertion struct {
		Service, EP, EPName, User string
		StepNumber                int
	}

	var idCounter int

	check := func(name, tmplStr string, expected []onCallAssertion) {
		t.Helper()

		users := make(map[string]*user.User)
		names := make(map[string]string)
		namesRev := make(map[string]string)

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			h := harness.NewHarness(t, sql, "escalation-policy-step-reorder")
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

			tmpl := template.New("mutation")
			tmpl.Funcs(template.FuncMap{
				"name": func(id string) string {
					if name, ok := names[id]; ok {
						return name
					}
					name := fmt.Sprintf("generated%d", idCounter)
					idCounter++
					names[id] = name
					namesRev[name] = id
					return name
				},
				"userID": func(id string) string {
					if usr, ok := users[id]; ok {
						return usr.ID
					}
					usr := h.CreateUser()
					users[id] = usr
					return usr.ID
				},
				"uuid": func(id string) string {
					return h.UUID(id)
				},
			})

			tmpl, err := tmpl.Parse(tmplStr)
			require.NoError(t, err)

			var buf bytes.Buffer
			err = tmpl.Execute(&buf, nil)
			require.NoError(t, err)

			query := buf.String()

			doQL(t, query, nil)
			h.Trigger()

			var onCallState map[string]struct {
				OnCallSteps []struct {
					StepNumber       int
					EscalationPolicy struct {
						Name       string
						AssignedTo []struct{ Name string }
					}
				}
			}

			buf.Reset()
			err = onCallAsnQueryTmpl.Execute(&buf, users)
			require.NoError(t, err, "render query")

			doQL(t, buf.String(), &onCallState)

			// map response to same type as expected value
			var actualOnCall []onCallAssertion
			for id, state := range onCallState {
				for _, step := range state.OnCallSteps {
					for _, svc := range step.EscalationPolicy.AssignedTo {
						ep := namesRev[step.EscalationPolicy.Name]
						var epName string

						if ep == "" {
							epName = step.EscalationPolicy.Name
						}

						actualOnCall = append(actualOnCall, onCallAssertion{
							User: id, StepNumber: step.StepNumber,
							Service: namesRev[svc.Name], EP: ep, EPName: epName,
						})
					}
				}
			}

			cpy := make([]onCallAssertion, len(expected))
			copy(cpy, expected) // copy so that we don't modify the original slice
			expected = cpy

			sortAssertions := func(slice []onCallAssertion) {
				sort.Slice(slice, func(a, b int) bool {
					if slice[a].User != slice[b].User {
						return slice[a].User < slice[b].User
					}
					if slice[a].Service != slice[b].Service {
						return slice[a].Service < slice[b].Service
					}
					if slice[a].EP != slice[b].EP {
						return slice[a].EP < slice[b].EP
					}
					if slice[a].StepNumber != slice[b].StepNumber {
						return slice[a].StepNumber < slice[b].StepNumber
					}
					return false
				})
			}
			sortAssertions(expected)
			sortAssertions(actualOnCall)

			if len(expected) == 0 {
				assert.Empty(t, actualOnCall)
				return
			}

			assert.EqualValues(t, expected, actualOnCall, "On-call assignments.")
		})
	}

	// User directly on EP is always on call
	check("User EP Direct", `
		mutation {
			createService(input:{
				name: "{{name  "svc"}}",
				newEscalationPolicy: {
					name: "{{name "ep"}}",
					steps: [{
						delayMinutes: 4,
						targets: [{type: user, id: "{{userID "bob"}}"}]
					}]
				}
			}){id}
		}
		`,
		[]onCallAssertion{
			{User: "bob", Service: "svc", EP: "ep"},
		},
	)

	// Active participant directly on EP is always on call

	check("User EP Rotation Direct", `
		mutation {
			createService(input:{
				name: "{{name  "svc"}}",
				newEscalationPolicy: {
					name: "{{name "ep"}}",
					steps: [{
						delayMinutes: 4,
						newRotation: {
							name: "{{name  "rot"}}",
							type: weekly,
							shiftLength: 1,
							timeZone: "UTC",
							start: "2006-01-02T15:04:05Z"
							userIDs: ["{{userID "bob"}}"]
						}
					}]
				}
			}){id}
		}`,
		[]onCallAssertion{
			{User: "bob", EP: "ep", Service: "svc"},
		},
	)

	check("Only One User EP Rotation Direct", `
		mutation {
			createService(input:{
				name: "{{name  "svc"}}",
				newEscalationPolicy: {
					name: "{{name "ep"}}",
					steps: [{
						delayMinutes: 4,
						newRotation: {
							name: "{{name  "rot1"}}",
							type: weekly,
							shiftLength: 1,
							timeZone: "UTC",
							start: "2006-01-02T15:04:05Z"
							userIDs: []
						}
					},{
						delayMinutes: 4,
						newRotation: {
							name: "{{name  "rot2"}}",
							type: weekly,
							shiftLength: 1,
							timeZone: "UTC",
							start: "2006-01-02T15:04:05Z"
							userIDs: ["{{userID "joe"}}"]
						}
					}]
				}
			}){id}
		}`,
		[]onCallAssertion{
			{Service: "svc", EP: "ep", StepNumber: 1, User: "joe"},
		},
	)

	// Different users on different rotations, users are on call but with different assignment rotations
	check("Multiple Users EP Rotation Direct", `
		mutation {
			createService(input:{
				name: "{{name  "svc"}}",
				newEscalationPolicy: {
					name: "{{name "ep"}}",
					steps: [{
						delayMinutes: 4,
						newRotation: {
							name: "{{name  "rot1"}}",
							type: weekly,
							shiftLength: 1,
							timeZone: "UTC",
							start: "2006-01-02T15:04:05Z"
							userIDs: ["{{userID "bob"}}"]
						}
					},{
						delayMinutes: 4,
						newRotation: {
							name: "{{name  "rot2"}}",
							type: weekly,
							shiftLength: 1,
							timeZone: "UTC",
							start: "2006-01-02T15:04:05Z"
							userIDs: ["{{userID "joe"}}"]
						}
					}]
				}
			}){id}
		}`,
		[]onCallAssertion{
			{Service: "svc", EP: "ep", StepNumber: 0, User: "bob"},
			{Service: "svc", EP: "ep", StepNumber: 1, User: "joe"},
		},
	)

	// EP -> Schedule, where there is an active ADD for a user
	check("User EP Schedule Add an Active Override", `
		mutation {
			createService(input:{
				name: "{{name  "svc"}}",
				newEscalationPolicy: {
					name: "{{name "ep"}}",
					steps: [{
								delayMinutes: 1
								newSchedule: {
									name: "{{name "sched"}}"
									timeZone: "UTC"
									newUserOverrides: [
										{
											addUserID: "{{userID "bob"}}"
											start: "1006-01-02T15:04:05Z"
											end: "4006-01-02T15:04:05Z"
										}
									]
								}
							}]
				}
			}){id}
		}`,
		[]onCallAssertion{
			{Service: "svc", EP: "ep", StepNumber: 0, User: "bob"},
		},
	)

	// EP -> Schedule, where there is an inactive ADD for a user
	check("User EP Schedule Inactive Add Override", `
		mutation {
			createService(input:{
				name: "{{name  "svc"}}",
				newEscalationPolicy: {
					name: "{{name "ep"}}",
					steps: [{
								delayMinutes: 1
								newSchedule: {
									name: "{{name "sched"}}"
									timeZone: "UTC"
									newUserOverrides: [
										{
											addUserID: "{{userID "sam"}}"
											start: "3006-01-02T15:04:05Z"
											end: "4006-01-02T15:04:05Z"
										}
									]
								}
							}]
				}
			}){id}
		}`,
		[]onCallAssertion{},
	)

	check("User EP Schedule Replace Override", `
		mutation {
			createService(input:{
				name: "{{name  "svc"}}",
				newEscalationPolicy: {
					name: "{{name "ep"}}",
					steps: [{
								delayMinutes: 1
								newSchedule: {
									name: "{{name "sched"}}"
									timeZone: "UTC"
									targets: [
									{
										target: {id:"{{userID "joe"}}", type:user}
										rules: [
											{
												start: "00:00"
												end: "23:59"
												weekdayFilter: [true, true, true, true, true, true, true]
											}
										]
									}
								]
									newUserOverrides: [
										{
											addUserID: "{{userID "bob"}}"
											removeUserID: "{{userID "joe"}}"
											start: "1006-01-02T15:04:05Z"
											end: "4006-01-02T15:04:05Z"
										}
									]
								}
							}]
				}
			}){id}
		}`,
		[]onCallAssertion{{Service: "svc", EP: "ep", StepNumber: 0, User: "bob"}},
	)

	//  Active schedule rule, user is replaced but in the future (inactive replacement)
	check("User EP Schedule Inactive Replace Override", `
		mutation {
			createService(input:{
				name: "{{name  "svc"}}",
				newEscalationPolicy: {
					name: "{{name "ep"}}",
					steps: [{
								delayMinutes: 1
								newSchedule: {
									name: "{{name "sched"}}"
									timeZone: "UTC"
									targets: [
									{
										target: {id:"{{userID "joe"}}", type:user}
										rules: [
											{
												start: "00:00"
												end: "23:59"
												weekdayFilter: [true, true, true, true, true, true, true]
											}
										]
									}
								]
									newUserOverrides: [
										{
											addUserID: "{{userID "bob"}}"
											removeUserID: "{{userID "joe"}}"
											start: "3006-01-02T15:04:05Z"
											end: "4006-01-02T15:04:05Z"
										}
									]
								}
							}]
				}
			}){id}
		}`,
		[]onCallAssertion{{Service: "svc", EP: "ep", StepNumber: 0, User: "joe"}},
	)

	// Same scenario, user is NOT replaced (no override)
	check("User EP Schedule Replace Override Absent", `
		mutation {
			createService(input:{
				name: "{{name  "svc"}}",
				newEscalationPolicy: {
					name: "{{name "ep"}}",
					steps: [{
								delayMinutes: 1
								newSchedule: {
									name: "{{name "sched"}}"
									timeZone: "UTC"
									targets: [
									{
										target: {id:"{{userID "joe"}}", type:user}
										rules: [
											{
												start: "00:00"
												end: "23:59"
												weekdayFilter: [true, true, true, true, true, true, true]
											}
										]
									}
								]
								}
							}]
				}
			}){id}
		}`,
		[]onCallAssertion{{Service: "svc", EP: "ep", StepNumber: 0, User: "joe"}},
	)

	// Same scenario, user is NOT replaced (no override), inactive schedule rule
	check("User EP Schedule No Days Replace Override Absent", `
		mutation {
			createService(input:{
				name: "{{name  "svc"}}",
				newEscalationPolicy: {
					name: "{{name "ep"}}",
					steps: [{
								delayMinutes: 1
								newSchedule: {
									name: "{{name "sched"}}"
									timeZone: "UTC"
									targets: [
									{
										target: {id:"{{userID "joe"}}", type:user}
										rules: [
											{
												start: "00:00"
												end: "23:59"
												weekdayFilter: [false, false, false, false, false, false, false]
											}
										]
									}
								]
								}
							}]
				}
			}){id}
		}`,
		[]onCallAssertion{},
	)

	// Active schedule rule, active rotation participant is replaced
	check("User EP Schedule Replace Rotation Override", `
		mutation {
		createService(
			input: {
				name: "{{name  "svc"}}"
				newEscalationPolicy: {
					name: "{{name  "ep"}}"
					steps: [
						{
							delayMinutes: 1
							newSchedule: {
								name: "{{name  "sched"}}"
								timeZone: "UTC"
								targets: [
									{
										newRotation: {
											name: "{{name  "rot"}}"
											type: weekly
											shiftLength: 1
											timeZone: "UTC"
											start: "2006-01-02T15:04:05Z"
											userIDs: ["{{userID "joe"}}"]
										}
										rules: [
											{
												start: "00:00"
												end: "23:59"
												weekdayFilter: [true, true, true, true, true, true, true]
											}
										]
									}
								]
								newUserOverrides: [
									{
										addUserID: "{{userID "bob"}}"
										removeUserID: "{{userID "joe"}}"
										start: "1006-01-02T15:04:05Z"
										end: "4006-01-02T15:04:05Z"
									}
								]
							}
						}
					]
				}
			}
		) {
			id
		}
	}`,
		[]onCallAssertion{{Service: "svc", EP: "ep", StepNumber: 0, User: "bob"}},
	)

	// Active schedule rule, active rotation participant is replaced with an inactive replace override
	check("User EP Schedule Replace Rotation Override (Inactive)", `
		mutation {
		createService(
			input: {
				name: "{{name  "svc"}}"
				newEscalationPolicy: {
					name: "{{name  "ep"}}"
					steps: [
						{
							delayMinutes: 1
							newSchedule: {
								name: "{{name  "sched"}}"
								timeZone: "UTC"
								targets: [
									{
										newRotation: {
											name: "{{name  "rot"}}"
											type: weekly
											shiftLength: 1
											timeZone: "UTC"
											start: "2006-01-02T15:04:05Z"
											userIDs: ["{{userID "joe"}}"]
										}
										rules: [
											{
												start: "00:00"
												end: "23:59"
												weekdayFilter: [true, true, true, true, true, true, true]
											}
										]
									}
								]
								newUserOverrides: [
									{
										addUserID: "{{userID "bob"}}"
										removeUserID: "{{userID "joe"}}"
										start: "3006-01-02T15:04:05Z"
										end: "4006-01-02T15:04:05Z"
									}
								]
							}
						}
					]
				}
			}
		) {
			id
		}
	}`,
		[]onCallAssertion{{Service: "svc", EP: "ep", StepNumber: 0, User: "joe"}},
	)

	// Same as above, but no service assignment
	check("User EP Schedule Replace Rotation Override No Service", `
		mutation {
		createEscalationPolicy(
			input: {
				name: "{{name  "ep"}}"
				steps: [
					{
						delayMinutes: 1
						newSchedule: {
							name: "{{name  "sched"}}"
							timeZone: "UTC"
							targets: [
								{
									newRotation: {
										name: "{{name  "rot"}}"
										type: weekly
										shiftLength: 1
										timeZone: "UTC"
										start: "2006-01-02T15:04:05Z"
										userIDs: ["{{userID "joe"}}"]
									}
									rules: [
										{
											start: "00:00"
											end: "23:59"
											weekdayFilter: [true, true, true, true, true, true, true]
										}
									]
								}
							]
							newUserOverrides: [
								{
									addUserID: "{{userID "bob"}}"
									removeUserID: "{{userID "joe"}}"
									start: "1006-01-02T15:04:05Z"
									end: "4006-01-02T15:04:05Z"
								}
							]
						}
					}
				]
			}
		) {
			id
		}
	}`,
		[]onCallAssertion{},
	)

	// User EP Schedule Replace Rotation Override Double Service
	check("User EP Schedule Replace Rotation Override Double Service", `
		mutation {
		alias0: createService(input: { name: "{{name "svc1"}}", escalationPolicyID: "{{uuid "eid"}}" }) {
			id
		}
		alias1: createService(input: { name: "{{name "svc2"}}", escalationPolicyID: "{{uuid "eid"}}" }) {
			id
		}
		createEscalationPolicyStep(
			input: {
				escalationPolicyID: "{{uuid "eid"}}"
				delayMinutes: 1
				newSchedule: {
					name: "{{name "sched"}}"
					timeZone: "UTC"
					targets: [
						{
							newRotation: {
								name: "{{name "rot"}}"
								type: weekly
								shiftLength: 1
								timeZone: "UTC"
								start: "2006-01-02T15:04:05Z"
								userIDs: ["{{userID "joe"}}"]
							}
							rules: [
								{
									start: "00:00"
									end: "23:59"
									weekdayFilter: [true, true, true, true, true, true, true]
								}
							]
						}
					]
					newUserOverrides: [
						{
							addUserID: "{{userID "bob"}}"
							removeUserID: "{{userID "joe"}}"
							start: "1006-01-02T15:04:05Z"
							end: "4006-01-02T15:04:05Z"
						}
					]
				}
			}
		) {
			id
		}
	}`,
		[]onCallAssertion{
			{Service: "svc1", EPName: "esc policy", StepNumber: 0, User: "bob"},
			{Service: "svc2", EPName: "esc policy", StepNumber: 0, User: "bob"},
		},
	)

	// Active schedule rule, active rotation participant is NOT replaced (no override)
	check("User EP Schedule Replace Rotation Override Absent", `
		mutation {
		createService(
			input: {
				name: "{{name  "svc"}}"
				newEscalationPolicy: {
					name: "{{name  "ep"}}"
					steps: [
						{
							delayMinutes: 1
							newSchedule: {
								name: "{{name  "sched"}}"
								timeZone: "UTC"
								targets: [
									{
										newRotation: {
											name: "{{name  "rot"}}"
											type: weekly
											shiftLength: 1
											timeZone: "UTC"
											start: "2006-01-02T15:04:05Z"
											userIDs: ["{{userID "joe"}}"]
										}
										rules: [
											{
												start: "00:00"
												end: "23:59"
												weekdayFilter: [true, true, true, true, true, true, true]
											}
										]
									}
								]
							}
						}
					]
				}
			}
		) {
			id
		}
	}`,
		[]onCallAssertion{{Service: "svc", EP: "ep", StepNumber: 0, User: "joe"}},
	)

	//	Active schedule rule, active rotation participant is removed
	check("User EP Schedule Remove Rotation Override", `
		mutation {
		createService(
			input: {
				name: "{{name  "svc"}}"
				newEscalationPolicy: {
					name: "{{name  "ep"}}"
					steps: [
						{
							delayMinutes: 1
							newSchedule: {
								name: "{{name  "sched"}}"
								timeZone: "UTC"
								targets: [
									{
										newRotation: {
											name: "{{name  "rot"}}"
											type: weekly
											shiftLength: 1
											timeZone: "UTC"
											start: "2006-01-02T15:04:05Z"
											userIDs: ["{{userID "joe"}}"]
										}
										rules: [
											{
												start: "00:00"
												end: "23:59"
												weekdayFilter: [true, true, true, true, true, true, true]
											}
										]
									}
								]
								newUserOverrides: [
									{
										removeUserID: "{{userID "joe"}}"
										start: "1006-01-02T15:04:05Z"
										end: "4006-01-02T15:04:05Z"
									}
								]
							}
						}
					]
				}
			}
		) {
			id
		}
	}`,
		[]onCallAssertion{},
	)

	// Active schedule rule, user is removed
	check("User EP Schedule Remove Override", `
		mutation {
		createService(
			input: {
				name: "{{name  "svc"}}"
				newEscalationPolicy: {
					name: "{{name  "ep"}}"
					steps: [
						{
							delayMinutes: 1
							newSchedule: {
								name: "{{name  "sched"}}"
								timeZone: "UTC"
								targets: [
									{
										target: {id:"{{userID "joe"}}", type:user}
										rules: [
											{
												start: "00:00"
												end: "23:59"
												weekdayFilter: [true, true, true, true, true, true, true]
											}
										]
									}
								]
								newUserOverrides: [
									{
										removeUserID: "{{userID "joe"}}"
										start: "1006-01-02T15:04:05Z"
										end: "4006-01-02T15:04:05Z"
									}
								]
							}
						}
					]
				}
			}
		) {
			id
		}
	}`,
		[]onCallAssertion{},
	)

	// Multiple add overrides, active schedule rules
	check("User EP Schedule Multiple Overrides", `
		mutation {
		createService(
			input: {
				name: "{{name  "svc"}}"
				newEscalationPolicy: {
					name: "{{name  "ep"}}"
					steps: [
						{
							delayMinutes: 1
							newSchedule: {
								name: "{{name  "sched"}}"
								timeZone: "UTC"
								targets: [
									{
										target: {id:"{{userID "joe"}}", type:user}
										rules: [
											{
												start: "00:00"
												end: "23:59"
												weekdayFilter: [true, true, true, true, true, true, true]
											}
										]
									}
								]
								newUserOverrides: [
									{
										addUserID: "{{userID "bob"}}"
										start: "1006-01-02T15:04:05Z"
										end: "4006-01-02T15:04:05Z"
									},
									{
										addUserID: "{{userID "joe"}}"
										start: "1006-01-02T15:04:05Z"
										end: "4006-01-02T15:04:05Z"
									}
								]
							}
						}
					]
				}
			}
		) {
			id
		}
	}`,
		[]onCallAssertion{
			{Service: "svc", EP: "ep", StepNumber: 0, User: "bob"},
			{Service: "svc", EP: "ep", StepNumber: 0, User: "joe"},
		},
	)

}
