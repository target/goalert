package graphql

import (
	"time"

	g "github.com/graphql-go/graphql"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/heartbeat"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/override"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/service"
	"github.com/target/goalert/util/timeutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

func parseUO(_m interface{}) (*override.UserOverride, error) {
	m, ok := _m.(map[string]interface{})
	if !ok {
		return nil, nil
	}
	var o override.UserOverride
	o.AddUserID, _ = m["add_user_id"].(string)
	o.RemoveUserID, _ = m["remove_user_id"].(string)
	o.Target = parseTarget(m)

	sTime, _ := m["start_time"].(string)
	var err error
	o.Start, err = time.Parse(time.RFC3339, sTime)
	if err != nil {
		return nil, validation.NewFieldError("start_time", "invalid format for time value: "+err.Error())
	}
	eTime, _ := m["end_time"].(string)
	o.End, err = time.Parse(time.RFC3339, eTime)
	if err != nil {
		return nil, validation.NewFieldError("end_time", "invalid format for time value: "+err.Error())
	}

	return &o, nil
}

func parseEP(_m interface{}) *escalation.Policy {
	m, ok := _m.(map[string]interface{})
	if !ok {
		return nil
	}

	var ep escalation.Policy
	ep.ID, _ = m["id_placeholder"].(string)
	ep.Name, _ = m["name"].(string)
	ep.Description, _ = m["description"].(string)
	ep.Repeat, _ = m["repeat"].(int)

	return &ep
}

func parseEPStep(_m interface{}) *escalation.Step {
	m, ok := _m.(map[string]interface{})
	if !ok {
		return nil
	}

	var step escalation.Step
	step.DelayMinutes, _ = m["delay_minutes"].(int)
	step.PolicyID, _ = m["escalation_policy_id"].(string)

	tgts, _ := m["targets"].([]interface{})
	for _, t := range tgts {
		tgt := parseTarget(t)
		if tgt == nil {
			continue
		}

		step.Targets = append(step.Targets, tgt)
	}

	return &step
}

func parseService(_m interface{}) *service.Service {
	m, ok := _m.(map[string]interface{})
	if !ok {
		return nil
	}

	var s service.Service
	s.ID, _ = m["id_placeholder"].(string)
	s.Name, _ = m["name"].(string)
	s.Description, _ = m["description"].(string)
	s.EscalationPolicyID, _ = m["escalation_policy_id"].(string)

	return &s
}

func parseHeartbeatMonitor(_m interface{}) *heartbeat.Monitor {
	m, ok := _m.(map[string]interface{})
	if !ok {
		return nil
	}

	var hb heartbeat.Monitor
	hb.Name, _ = m["name"].(string)
	min, _ := m["interval_minutes"].(int)
	hb.Timeout = time.Duration(min) * time.Minute
	hb.ServiceID, _ = m["service_id"].(string)

	return &hb
}

func parseIntegrationKey(_m interface{}) *integrationkey.IntegrationKey {
	m, ok := _m.(map[string]interface{})
	if !ok {
		return nil
	}

	var key integrationkey.IntegrationKey
	key.Name, _ = m["name"].(string)
	key.Type, _ = m["type"].(integrationkey.Type)
	key.ServiceID, _ = m["service_id"].(string)

	return &key
}

func parseRotation(_m interface{}) (*rotation.Rotation, error) {
	m, ok := _m.(map[string]interface{})
	if !ok {
		return nil, validation.NewFieldError("input", "invalid input object")
	}

	var rot rotation.Rotation
	rot.ID, _ = m["id_placeholder"].(string)
	rot.Name, _ = m["name"].(string)
	rot.Description, _ = m["description"].(string)
	sTime, _ := m["start"].(string)
	var err error
	rot.Start, err = time.Parse(time.RFC3339, sTime)
	if err != nil {
		return nil, validation.NewFieldError("start", "invalid format for time value: "+err.Error())
	}
	tz, _ := m["time_zone"].(string)
	if tz == "" {
		return nil, validation.NewFieldError("time_zone", "must not be empty")
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, validation.NewFieldError("time_zone", err.Error())
	}
	rot.Start = rot.Start.In(loc)
	rot.Type = m["type"].(rotation.Type)
	rot.ShiftLength = m["shift_length"].(int)

	return &rot, nil
}

func parseRotationPart(_m interface{}) (*rotation.Participant, error) {
	m, ok := _m.(map[string]interface{})
	if !ok {
		return nil, validation.NewFieldError("input", "invalid input object")
	}

	var rp rotation.Participant
	rp.RotationID, _ = m["rotation_id"].(string)
	rp.Target = assignment.UserTarget(m["user_id"].(string))

	return &rp, nil
}

func parseSched(_m interface{}) (*schedule.Schedule, error) {
	m, ok := _m.(map[string]interface{})
	if !ok {
		return nil, validation.NewFieldError("input", "invalid input object")
	}

	var s schedule.Schedule
	s.ID, _ = m["id_placeholder"].(string)
	s.Name, _ = m["name"].(string)
	s.Description, _ = m["description"].(string)

	tz, _ := m["time_zone"].(string)
	if tz == "" {
		return nil, validation.NewFieldError("time_zone", "must not be empty")
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, validation.NewFieldError("time_zone", err.Error())
	}

	s.TimeZone = loc

	return &s, nil
}

func parseSchedRule(_m interface{}) (*rule.Rule, error) {
	m, ok := _m.(map[string]interface{})
	if !ok {
		return nil, validation.NewFieldError("input", "invalid input object")
	}

	var r rule.Rule
	r.ScheduleID = m["schedule_id"].(string)

	var e bool
	e, _ = m["sunday"].(bool)
	r.SetDay(time.Sunday, e)
	e, _ = m["monday"].(bool)
	r.SetDay(time.Monday, e)
	e, _ = m["tuesday"].(bool)
	r.SetDay(time.Tuesday, e)
	e, _ = m["wednesday"].(bool)
	r.SetDay(time.Wednesday, e)
	e, _ = m["thursday"].(bool)
	r.SetDay(time.Thursday, e)
	e, _ = m["friday"].(bool)
	r.SetDay(time.Friday, e)
	e, _ = m["saturday"].(bool)
	r.SetDay(time.Saturday, e)

	startStr, _ := m["start"].(string)
	endStr, _ := m["end"].(string)
	var err error
	r.Start, err = timeutil.ParseClock(startStr)
	if err != nil {
		return nil, validation.NewFieldError("start", err.Error())
	}
	r.End, err = timeutil.ParseClock(endStr)
	if err != nil {
		return nil, validation.NewFieldError("end", err.Error())
	}

	r.Target = parseTarget(m["target"])

	return &r, nil
}

func parseTarget(_m interface{}) assignment.Target {
	m, ok := _m.(map[string]interface{})
	if !ok {
		return nil
	}
	var raw assignment.RawTarget
	raw.ID, _ = m["target_id"].(string)
	raw.Type, _ = m["target_type"].(assignment.TargetType)

	return &raw
}

/*
 * Creates a service, userTarget, rotation, or schedule, an escalation policy,
 * and adds a step from the user, rot, or sched created. finally, generates an
 * integration key to return
 */
func (h *Handler) createAllField() *g.Field {
	return &g.Field{
		Type: h.createAll,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewNonNull(g.NewInputObject(g.InputObjectConfig{
					Name: "CreateAllInput",
					Description: "Creates up to any number of escalation policies, steps, services, integration keys," +
						"rotations, participants, schedules, and schedule rules.",
					Fields: g.InputObjectConfigFieldMap{
						"escalation_policies": &g.InputObjectFieldConfig{
							Type: g.NewList(g.NewInputObject(g.InputObjectConfig{
								Name: "CreateAllEscalationPolicyInput",
								Fields: g.InputObjectConfigFieldMap{
									"id_placeholder": &g.InputObjectFieldConfig{Type: g.String},
									"name":           &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
									"description":    &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
									"repeat":         &g.InputObjectFieldConfig{Type: g.Int},
								},
							})),
						},
						"escalation_policy_steps": &g.InputObjectFieldConfig{
							Type: g.NewList(g.NewInputObject(g.InputObjectConfig{
								Name: "CreateAllEscalationPolicyStepInput",
								Fields: g.InputObjectConfigFieldMap{
									"escalation_policy_id": &g.InputObjectFieldConfig{
										Description: "The UUID of an existing policy or the value of an id_placeholder from the current request.",
										Type:        g.NewNonNull(g.String),
									},
									"delay_minutes": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Int)},
									"targets": &g.InputObjectFieldConfig{
										Type: g.NewList(g.NewInputObject(g.InputObjectConfig{
											Name: "CreateAllEPStepTargetInput",
											Fields: g.InputObjectConfigFieldMap{
												"target_id":   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
												"target_type": &g.InputObjectFieldConfig{Type: g.NewNonNull(epStepTarget)},
											},
										})),
									},
								}})),
						},
						"services": &g.InputObjectFieldConfig{
							Type: g.NewList(g.NewInputObject(g.InputObjectConfig{
								Name: "CreateAllServiceInput",
								Fields: g.InputObjectConfigFieldMap{
									"escalation_policy_id": &g.InputObjectFieldConfig{
										Description: "The UUID of an existing policy or the value of an id_placeholder from the current request.",
										Type:        g.NewNonNull(g.String),
									},
									"id_placeholder": &g.InputObjectFieldConfig{Type: g.String},
									"name":           &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
									"description":    &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
								},
							})),
						},
						"integration_keys": &g.InputObjectFieldConfig{
							Type: g.NewList(g.NewInputObject(g.InputObjectConfig{
								Name: "CreateAllIntegrationKeyInput",
								Fields: g.InputObjectConfigFieldMap{
									"service_id": &g.InputObjectFieldConfig{
										Description: "The UUID of an existing service or the value of an id_placeholder from the current request.",
										Type:        g.NewNonNull(g.String),
									},
									"name": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
									"type": &g.InputObjectFieldConfig{Type: g.NewNonNull(integrationKeyType)},
								},
							})),
						},
						"heartbeat_monitors": &g.InputObjectFieldConfig{
							Type: g.NewList(g.NewInputObject(g.InputObjectConfig{
								Name: "CreateAllHeartbeatMonitorInput",
								Fields: g.InputObjectConfigFieldMap{
									"service_id": &g.InputObjectFieldConfig{
										Description: "The UUID of an existing service or the value of an id_placeholder from the current request.",
										Type:        g.NewNonNull(g.String),
									},
									"name":             &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
									"interval_minutes": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Int)},
								},
							})),
						},
						"rotations": &g.InputObjectFieldConfig{
							Type: g.NewList(g.NewInputObject(g.InputObjectConfig{
								Name: "CreateAllRotationInput",
								Fields: g.InputObjectConfigFieldMap{
									"id_placeholder": &g.InputObjectFieldConfig{Type: g.String},
									"name":           &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
									"description":    &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
									"time_zone":      &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
									"type":           &g.InputObjectFieldConfig{Type: g.NewNonNull(rotationTypeEnum)},
									"start":          &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
									"shift_length":   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Int)},
								},
							})),
						},
						"rotation_participants": &g.InputObjectFieldConfig{
							Type: g.NewList(g.NewInputObject(g.InputObjectConfig{
								Name: "CreateAllRotationParticipantInput",
								Fields: g.InputObjectConfigFieldMap{
									"rotation_id": &g.InputObjectFieldConfig{
										Description: "The UUID of an existing rotation or the value of an id_placeholder from the current request.",
										Type:        g.NewNonNull(g.String),
									},
									"user_id": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
								},
							})),
						},
						"schedules": &g.InputObjectFieldConfig{
							Type: g.NewList(g.NewInputObject(g.InputObjectConfig{
								Name: "CreateAllScheduleInput",
								Fields: g.InputObjectConfigFieldMap{
									"id_placeholder": &g.InputObjectFieldConfig{Type: g.String},
									"name":           &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
									"description":    &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
									"time_zone":      &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
								},
							})),
						},
						"user_overrides": &g.InputObjectFieldConfig{
							Type: g.NewList(g.NewInputObject(g.InputObjectConfig{
								Name: "CreateAllUserOverrideInput",
								Fields: g.InputObjectConfigFieldMap{
									"target_id":      &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
									"target_type":    &g.InputObjectFieldConfig{Type: g.NewNonNull(userOverrideTargetType)},
									"start_time":     &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
									"end_time":       &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
									"add_user_id":    &g.InputObjectFieldConfig{Type: g.String},
									"remove_user_id": &g.InputObjectFieldConfig{Type: g.String},
								},
							})),
						},
						"schedule_rules": &g.InputObjectFieldConfig{
							Type: g.NewList(g.NewInputObject(g.InputObjectConfig{
								Name: "CreateAllScheduleRuleInput",
								Fields: g.InputObjectConfigFieldMap{
									"schedule_id": &g.InputObjectFieldConfig{
										Description: "The UUID of an existing schedule or the value of an id_placeholder from the current request.",
										Type:        g.NewNonNull(g.String),
									},
									"sunday":    &g.InputObjectFieldConfig{Type: g.Boolean},
									"monday":    &g.InputObjectFieldConfig{Type: g.Boolean},
									"tuesday":   &g.InputObjectFieldConfig{Type: g.Boolean},
									"wednesday": &g.InputObjectFieldConfig{Type: g.Boolean},
									"thursday":  &g.InputObjectFieldConfig{Type: g.Boolean},
									"friday":    &g.InputObjectFieldConfig{Type: g.Boolean},
									"saturday":  &g.InputObjectFieldConfig{Type: g.Boolean},
									"start":     &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
									"end":       &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
									"target": &g.InputObjectFieldConfig{
										Type: g.NewNonNull(g.NewInputObject(g.InputObjectConfig{
											Name: "CreateAllScheduleRuleTargetInput",
											Fields: g.InputObjectConfigFieldMap{
												"target_id":   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
												"target_type": &g.InputObjectFieldConfig{Type: g.NewNonNull(schedRuleTarget)},
											},
										})),
									},
								}})),
						},
					},
				})),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, validation.NewFieldError("input", "expected object")
			}

			scrub := newScrubber(p.Context).scrub
			const limitTotal = 35
			var count int
			var err error

			// parse everything
			getSlice := func(s string) []interface{} {
				v, _ := m[s].([]interface{})

				count += len(v)
				if count > limitTotal {
					v = v[:0]
					err = validate.Many(err, validation.NewFieldError(s, "too many items"))
				}

				return v
			}

			var data createAllData

			for _, v := range getSlice("escalation_policies") {
				ep := parseEP(v)
				if ep != nil {
					data.EscalationPolicies = append(data.EscalationPolicies, *ep)
				}
			}

			for _, v := range getSlice("escalation_policy_steps") {
				step := parseEPStep(v)
				if step != nil {
					data.EscalationPolicySteps = append(data.EscalationPolicySteps, *step)
				}
			}

			for _, v := range getSlice("services") {
				serv := parseService(v)
				if serv != nil {
					data.Services = append(data.Services, *serv)
				}
			}

			for _, v := range getSlice("integration_keys") {
				key := parseIntegrationKey(v)
				if key != nil {
					data.IntegrationKeys = append(data.IntegrationKeys, *key)
				}
			}

			for _, v := range getSlice("rotations") {
				rot, err := parseRotation(v)
				if err != nil {
					return scrub(nil, err)
				}
				if rot != nil {
					data.Rotations = append(data.Rotations, *rot)
				}
			}

			for _, v := range getSlice("rotation_participants") {
				rp, err := parseRotationPart(v)
				if err != nil {
					return scrub(nil, err)
				}
				if rp != nil {
					data.RotationParticipants = append(data.RotationParticipants, *rp)
				}
			}

			for _, v := range getSlice("schedules") {
				sched, err := parseSched(v)
				if err != nil {
					return scrub(nil, err)
				}
				if sched != nil {
					data.Schedules = append(data.Schedules, *sched)
				}
			}

			for _, v := range getSlice("user_overrides") {
				o, err := parseUO(v)
				if err != nil {
					return scrub(nil, err)
				}
				if o != nil {
					data.UserOverrides = append(data.UserOverrides, *o)
				}
			}

			for _, v := range getSlice("schedule_rules") {
				r, err := parseSchedRule(v)
				if err != nil {
					return scrub(nil, err)
				}
				if r != nil {
					data.ScheduleRules = append(data.ScheduleRules, *r)
				}
			}

			for _, v := range getSlice("heartbeat_monitors") {
				r := parseHeartbeatMonitor(v)
				if r != nil {
					data.HeartbeatMonitors = append(data.HeartbeatMonitors, *r)
				}
			}

			if err != nil {
				return nil, err
			}

			// create & return everything
			return scrub(h.c.createAll(p.Context, &data))
		},
	}
}

func (h *Handler) createAllFields() g.Fields {
	return g.Fields{
		"escalation_policies":     &g.Field{Type: g.NewList(h.escalationPolicy)},
		"escalation_policy_steps": &g.Field{Type: g.NewList(h.escalationPolicyStep)},
		"services":                &g.Field{Type: g.NewList(h.service)},
		"integration_keys":        &g.Field{Type: g.NewList(h.integrationKey)},
		"rotations":               &g.Field{Type: g.NewList(h.rotation)},
		"rotation_participants":   &g.Field{Type: g.NewList(h.rotationParticipant)},
		"schedules":               &g.Field{Type: g.NewList(h.schedule)},
		"schedule_rules":          &g.Field{Type: g.NewList(h.scheduleRule)},
		"heartbeat_monitors":      &g.Field{Type: g.NewList(h.heartbeat)},
		"user_overrides":          &g.Field{Type: g.NewList(h.userOverride)},
	}
}
