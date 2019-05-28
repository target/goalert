package graphql

import (
	"fmt"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation"
	"time"

	g "github.com/graphql-go/graphql"
	"github.com/pkg/errors"
)

func getSchedule(src interface{}) (*schedule.Schedule, error) {
	switch t := src.(type) {
	case schedule.Schedule:
		return &t, nil
	case *schedule.Schedule:
		return t, nil
	default:
		return nil, fmt.Errorf("invalid source type for schedule %T", t)
	}
}

func (h *Handler) getRotation(p g.ResolveParams) (r rotation.Rotation, err error) {
	s, err := getSchedule(p.Source)
	if err != nil {
		return r, err
	}

	rotID, err := h.legacyDB.RotationIDFromScheduleID(p.Context, s.ID)
	if err != nil {
		return r, errors.Wrap(err, "fetch rotation ID")
	}

	rt, err := h.c.RotationStore.FindRotation(p.Context, rotID)
	if err != nil {
		return r, errors.Wrap(err, "fetch rotation")
	}

	return *rt, nil
}

func (h *Handler) getRotations(p g.ResolveParams) ([]rotation.Rotation, error) {
	s, err := getSchedule(p.Source)
	if err != nil {
		return nil, err
	}
	ids, err := h.legacyDB.FindAllRotationIDsFromScheduleID(p.Context, s.ID)
	if err != nil {
		return nil, err
	}

	var result []rotation.Rotation
	for _, id := range ids {

		r, err := h.c.RotationStore.FindRotation(p.Context, id)
		if err != nil {
			return nil, err
		}
		result = append(result, *r)
	}

	return result, nil
}

func (h *Handler) scheduleFields() g.Fields {
	return g.Fields{
		"id":          &g.Field{Type: g.String},
		"name":        &g.Field{Type: g.String},
		"description": &g.Field{Type: g.String},
		"time_zone":   &g.Field{Type: g.String},
		"target_type": targetTypeField(assignment.TargetTypeSchedule),
		"escalation_policies": &g.Field{
			Type:        g.NewList(h.escalationPolicy),
			Description: "List of escalation policies currently using this schedule",
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				s, err := getSchedule(p.Source)
				if err != nil {
					return nil, err
				}
				scrub := newScrubber(p.Context).scrub

				return scrub(h.c.EscalationStore.FindAllPoliciesBySchedule(p.Context, s.ID))
			},
		},

		"user_overrides": &g.Field{
			Type: g.NewList(h.userOverride),
			Args: g.FieldConfigArgument{
				"start_time": &g.ArgumentConfig{Type: g.NewNonNull(g.String)},
				"end_time":   &g.ArgumentConfig{Type: g.NewNonNull(g.String)},
			},
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				s, err := getSchedule(p.Source)
				if err != nil {
					return nil, err
				}
				scrub := newScrubber(p.Context).scrub
				startStr, _ := p.Args["start_time"].(string)
				endStr, _ := p.Args["end_time"].(string)

				start, err := time.Parse(time.RFC3339, startStr)
				if err != nil {
					return nil, validation.NewFieldError("start_time", err.Error())
				}
				end, err := time.Parse(time.RFC3339, endStr)
				if err != nil {
					return nil, validation.NewFieldError("end_time", err.Error())
				}

				return scrub(h.c.OverrideStore.FindAllUserOverrides(p.Context, start, end, assignment.ScheduleTarget(s.ID)))
			},
		},

		"rotations": &g.Field{
			Type:    g.NewList(h.rotation),
			Resolve: func(p g.ResolveParams) (interface{}, error) { return newScrubber(p.Context).scrub(h.getRotations(p)) },
		},

		// rotation stuff
		"type": &g.Field{
			Type:              rotationTypeEnum,
			DeprecationReason: "Use the 'rotations' field.",
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				r, err := h.getRotation(p)
				return newScrubber(p.Context).scrub(r.Type, err)
			},
		},

		"start_time": &g.Field{
			Type:              ISOTimestamp,
			DeprecationReason: "Use the 'rotations' field.",
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				r, err := h.getRotation(p)
				return newScrubber(p.Context).scrub(r.Start, err)
			},
		},

		"shift_length": &g.Field{
			Type:              g.Int,
			DeprecationReason: "Use the 'rotations' field.",
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				r, err := h.getRotation(p)
				return newScrubber(p.Context).scrub(r.ShiftLength, err)
			},
		},

		"assignments": &g.Field{
			Type: g.NewList(h.scheduleAssignment),
			Args: g.FieldConfigArgument{
				"start_time": &g.ArgumentConfig{Type: g.NewNonNull(g.String)},
				"end_time":   &g.ArgumentConfig{Type: g.NewNonNull(g.String)},
			},
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				sched, err := getSchedule(p.Source)
				if err != nil {
					return nil, err
				}

				scrub := newScrubber(p.Context).scrub

				startStr, _ := p.Args["start_time"].(string)
				endStr, _ := p.Args["end_time"].(string)

				start, err := time.Parse(time.RFC3339, startStr)
				if err != nil {
					return nil, validation.NewFieldError("start_time", err.Error())
				}
				end, err := time.Parse(time.RFC3339, endStr)
				if err != nil {
					return nil, validation.NewFieldError("end_time", err.Error())
				}
				return scrub(h.c.ShiftCalc.ScheduleAssignments(p.Context, start, end, sched.ID))
			},
		},

		"final_shifts": &g.Field{
			Type: g.NewList(h.scheduleShift),
			Args: g.FieldConfigArgument{
				"start_time": &g.ArgumentConfig{Type: g.NewNonNull(g.String)},
				"end_time":   &g.ArgumentConfig{Type: g.NewNonNull(g.String)},
			},
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				sched, err := getSchedule(p.Source)
				if err != nil {
					return nil, err
				}

				scrub := newScrubber(p.Context).scrub

				startStr, _ := p.Args["start_time"].(string)
				endStr, _ := p.Args["end_time"].(string)

				start, err := time.Parse(time.RFC3339, startStr)
				if err != nil {
					return nil, validation.NewFieldError("start_time", err.Error())
				}
				end, err := time.Parse(time.RFC3339, endStr)
				if err != nil {
					return nil, validation.NewFieldError("end_time", err.Error())
				}
				return scrub(h.c.ShiftCalc.ScheduleFinalShifts(p.Context, start, end, sched.ID))
			},
		},

		"final_shifts_with_overrides": &g.Field{
			Type: g.NewList(h.scheduleShift),
			Args: g.FieldConfigArgument{
				"start_time": &g.ArgumentConfig{Type: g.NewNonNull(g.String)},
				"end_time":   &g.ArgumentConfig{Type: g.NewNonNull(g.String)},
				"v2":         &g.ArgumentConfig{Type: g.Boolean},
			},
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				sched, err := getSchedule(p.Source)
				if err != nil {
					return nil, err
				}

				scrub := newScrubber(p.Context).scrub

				startStr, _ := p.Args["start_time"].(string)
				endStr, _ := p.Args["end_time"].(string)

				start, err := time.Parse(time.RFC3339, startStr)
				if err != nil {
					return nil, validation.NewFieldError("start_time", err.Error())
				}
				end, err := time.Parse(time.RFC3339, endStr)
				if err != nil {
					return nil, validation.NewFieldError("end_time", err.Error())
				}

				if end.After(start.In(time.UTC).AddDate(0, 1, 5)) {
					return nil, validation.NewFieldError("end_time", "must not be more than 1 month beyond start_time")
				}
				if !end.After(start) {
					return nil, validation.NewFieldError("end_time", "must be after start_time")
				}

				if v2, _ := p.Args["v2"].(bool); v2 {
					return h.c.OnCallStore.HistoryBySchedule(p.Context, sched.ID, start, end)
				}
				return scrub(h.c.ShiftCalc.ScheduleFinalShiftsWithOverrides(p.Context, start, end, sched.ID))
			},
		},
	}
}

func (h *Handler) scheduleField() *g.Field {
	return &g.Field{
		Type: h.schedule,
		Args: g.FieldConfigArgument{
			"id": &g.ArgumentConfig{
				Type: g.NewNonNull(g.String),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			id, _ := p.Args["id"].(string)
			return newScrubber(p.Context).scrub(h.c.ScheduleStore.FindOne(p.Context, id))
		},
	}
}
func (h *Handler) schedulesField() *g.Field {
	return &g.Field{
		Type: g.NewList(h.schedule),
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			return newScrubber(p.Context).scrub(h.c.ScheduleStore.FindAll(p.Context))
		},
	}
}

func (h *Handler) createScheduleField() *g.Field {
	return &g.Field{
		Type: h.schedule,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewNonNull(g.NewInputObject(g.InputObjectConfig{
					Name:        "CreateScheduleInput",
					Description: "Create a schedule.",
					Fields: g.InputObjectConfigFieldMap{
						"name":        &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"description": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"time_zone":   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"default_rotation": &g.InputObjectFieldConfig{Type: g.NewInputObject(g.InputObjectConfig{
							Name: "DefaultRotationFields",
							Fields: g.InputObjectConfigFieldMap{
								"type":         &g.InputObjectFieldConfig{Type: g.NewNonNull(rotationTypeEnum)},
								"start_time":   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
								"shift_length": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Int)},
							},
						})},
					},
				})),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input")
			}
			scrub := newScrubber(p.Context).scrub

			var s schedule.Schedule
			var r rotation.Rotation
			s.Name, _ = m["name"].(string)
			s.Description, _ = m["description"].(string)
			z, _ := m["time_zone"].(string)

			var err error
			s.TimeZone, err = util.LoadLocation(z)
			if err != nil {
				return scrub(nil, validation.NewFieldError("Timezone", err.Error()))
			}

			rot, ok := m["default_rotation"].(map[string]interface{})
			if !ok {
				// no default, just create it
				return scrub(h.c.ScheduleStore.Create(p.Context, &s))
			}

			// Creating default rotation
			r.Name = s.Name + " Rotation"
			r.ShiftLength, _ = rot["shift_length"].(int)
			r.Type, _ = rot["type"].(rotation.Type)

			tz, _ := rot["time_zone"].(string)
			if tz == "" {
				tz = s.TimeZone.String()
			}
			loc, err := util.LoadLocation(tz)
			if err != nil {
				return nil, validation.NewFieldError("time_zone", err.Error())
			}

			sTime, _ := rot["start_time"].(string)
			r.Start, err = time.Parse(time.RFC3339, sTime)
			if err != nil {
				return nil, validation.NewFieldError("start_time", err.Error())
			}
			r.Start = r.Start.In(loc)
			tx, err := h.c.DB.BeginTx(p.Context, nil)
			if err != nil {
				return scrub(nil, err)
			}
			defer tx.Rollback()

			// need to create a rotation, a schedule, a rule, and then point the rule to the rotation
			newSched, err := h.c.ScheduleStore.CreateScheduleTx(p.Context, tx, &s)
			if err != nil {
				return scrub(nil, errors.Wrap(err, "create schedule"))
			}

			newRot, err := h.c.RotationStore.CreateRotationTx(p.Context, tx, &r)
			if err != nil {
				return scrub(nil, errors.Wrap(err, "create rotation for new schedule"))
			}

			_, err = h.c.ScheduleRuleStore.CreateRuleTx(p.Context, tx, rule.NewAlwaysActive(newSched.ID, assignment.RotationTarget(newRot.ID)))
			if err != nil {
				return scrub(nil, errors.Wrap(err, "create rule for new schedule and rotation"))
			}

			err = tx.Commit()
			if err != nil {
				return scrub(nil, err)
			}

			return newSched, nil
		},
	}
}

func (h *Handler) updateSchedule() *g.Field {
	return &g.Field{
		Type: h.schedule,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewNonNull(g.NewInputObject(g.InputObjectConfig{
					Name:        "UpdateScheduleInput",
					Description: "Update a schedule.",
					Fields: g.InputObjectConfigFieldMap{
						"id":          &g.InputObjectFieldConfig{Type: g.String, Description: "Specifies an existing schedule to update."},
						"name":        &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"description": &g.InputObjectFieldConfig{Type: g.String},
						"time_zone":   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
					},
				})),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input")
			}
			var s schedule.Schedule
			s.ID, _ = m["id"].(string)
			s.Name, _ = m["name"].(string)
			s.Description, _ = m["description"].(string)
			z, _ := m["time_zone"].(string)

			var err error
			s.TimeZone, err = util.LoadLocation(z)
			if err != nil {
				return newScrubber(p.Context).scrub(nil, errors.Wrap(err, "parse time_zone"))
			}

			err = h.c.ScheduleStore.Update(p.Context, &s)
			if err != nil {
				return newScrubber(p.Context).scrub(nil, errors.Wrap(err, "update schedule"))
			}
			return s, nil
		},
	}
}

func (h *Handler) deleteScheduleField() *g.Field {
	return &g.Field{
		Type: g.NewObject(g.ObjectConfig{
			Name:   "DeleteScheduleOutput",
			Fields: g.Fields{"deleted_id": &g.Field{Type: g.String}},
		}),
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewNonNull(g.NewInputObject(g.InputObjectConfig{
					Name: "DeleteScheduleInput",
					Fields: g.InputObjectConfigFieldMap{
						"id": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
					},
				})),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			var r struct {
				ID string `json:"deleted_id"`
			}
			r.ID, _ = m["id"].(string)
			return newScrubber(p.Context).scrub(r, h.c.ScheduleStore.Delete(p.Context, r.ID))
		},
	}
}
