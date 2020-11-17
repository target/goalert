package graphql

import (
	"fmt"
	"time"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/timeutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"

	g "github.com/graphql-go/graphql"
	"github.com/pkg/errors"
)

// type Rule struct {
// 	ID         string
// 	ScheduleID string
// 	Days       [7]bool
// 	Start      time.Time
// 	End        time.Time
// }
func getScheduleRule(src interface{}) (*rule.Rule, error) {
	switch r := src.(type) {
	case *rule.Rule:
		return r, nil
	case rule.Rule:
		return &r, nil
	default:
		return nil, fmt.Errorf("could not id of user (unknown source type %T)", r)
	}
}
func resolveDay(n time.Weekday) g.FieldResolveFn {
	return func(p g.ResolveParams) (interface{}, error) {
		rule, err := getScheduleRule(p.Source)
		if err != nil {
			return nil, err
		}
		return rule.Day(n), nil
	}
}
func (h *Handler) scheduleRuleFields() g.Fields {
	return g.Fields{
		"id": &g.Field{Type: g.String},

		"sunday":    &g.Field{Type: g.Boolean, Resolve: resolveDay(time.Sunday)},
		"monday":    &g.Field{Type: g.Boolean, Resolve: resolveDay(time.Monday)},
		"tuesday":   &g.Field{Type: g.Boolean, Resolve: resolveDay(time.Tuesday)},
		"wednesday": &g.Field{Type: g.Boolean, Resolve: resolveDay(time.Wednesday)},
		"thursday":  &g.Field{Type: g.Boolean, Resolve: resolveDay(time.Thursday)},
		"friday":    &g.Field{Type: g.Boolean, Resolve: resolveDay(time.Friday)},
		"saturday":  &g.Field{Type: g.Boolean, Resolve: resolveDay(time.Saturday)},

		"start": &g.Field{Type: HourTime},
		"end":   &g.Field{Type: HourTime},

		"summary": &g.Field{
			Type: g.String,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				r, err := getScheduleRule(p.Source)
				if err != nil {
					return nil, err
				}

				return r.String(), nil
			},
		},
	}
}

var schedRuleTarget = g.NewEnum(g.EnumConfig{
	Name: "ScheduleRuleTarget",
	Values: g.EnumValueConfigMap{
		"user":     &g.EnumValueConfig{Value: assignment.TargetTypeUser},
		"rotation": &g.EnumValueConfig{Value: assignment.TargetTypeRotation},
	},
})

func (h *Handler) createScheduleRuleField() *g.Field {
	return &g.Field{
		Type: h.schedule,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "CreateScheduleRuleInput",
					Fields: g.InputObjectConfigFieldMap{
						"schedule_id": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},

						"sunday":    &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Boolean)},
						"monday":    &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Boolean)},
						"tuesday":   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Boolean)},
						"wednesday": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Boolean)},
						"thursday":  &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Boolean)},
						"friday":    &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Boolean)},
						"saturday":  &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Boolean)},

						"start": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"end":   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},

						"target_type": &g.InputObjectFieldConfig{Type: g.NewNonNull(schedRuleTarget)},
						"target_id":   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			var r rule.Rule
			r.ScheduleID, _ = m["schedule_id"].(string)
			err := validate.UUID("schedule_id", r.ScheduleID)
			if err != nil {
				return nil, err
			}

			var asnTarget assignment.RawTarget
			asnTarget.Type, _ = m["target_type"].(assignment.TargetType)
			asnTarget.ID, _ = m["target_id"].(string)
			err = validate.UUID("target_id", asnTarget.ID)
			if err != nil {
				return nil, err
			}

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
			r.Start, err = timeutil.ParseClock(startStr)
			if err != nil {
				return nil, validation.NewFieldError("start", err.Error())
			}
			r.End, err = timeutil.ParseClock(endStr)
			if err != nil {
				return nil, validation.NewFieldError("end", err.Error())
			}
			r.Target = asnTarget

			scrub := newScrubber(p.Context).scrub

			_, err = h.c.ScheduleRuleStore.Add(p.Context, &r)
			if err != nil {
				return scrub(nil, err)
			}

			return scrub(h.c.ScheduleStore.FindOne(p.Context, r.ScheduleID))
		},
	}
}

func (h *Handler) updateScheduleRuleField() *g.Field {
	return &g.Field{
		Type: h.schedule,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "UpdateScheduleRuleInput",
					Fields: g.InputObjectConfigFieldMap{
						"id": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},

						"sunday":    &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Boolean)},
						"monday":    &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Boolean)},
						"tuesday":   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Boolean)},
						"wednesday": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Boolean)},
						"thursday":  &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Boolean)},
						"friday":    &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Boolean)},
						"saturday":  &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Boolean)},

						"start": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"end":   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			var r rule.Rule
			r.ID, _ = m["id"].(string)
			err := validate.UUID("id", r.ID)
			if err != nil {
				return nil, err
			}
			p.Context = log.WithField(p.Context, "ScheduleRuleID", r.ID)
			scrub := newScrubber(p.Context).scrub
			oldRule, err := h.c.ScheduleRuleStore.FindOne(p.Context, r.ID)
			if err != nil {
				return scrub(nil, err)
			}
			r.ScheduleID = oldRule.ScheduleID
			r.Target = oldRule.Target

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
			r.Start, err = timeutil.ParseClock(startStr)
			if err != nil {
				return nil, validation.NewFieldError("start", err.Error())
			}
			r.End, err = timeutil.ParseClock(endStr)
			if err != nil {
				return nil, validation.NewFieldError("end", err.Error())
			}

			err = h.c.ScheduleRuleStore.Update(p.Context, &r)
			if err != nil {
				return scrub(nil, err)
			}

			return scrub(h.c.ScheduleStore.FindOne(p.Context, r.ScheduleID))
		},
	}
}

func (h *Handler) deleteScheduleRuleField() *g.Field {
	return &g.Field{
		Type: h.schedule,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "DeleteScheduleRuleInput",
					Fields: g.InputObjectConfigFieldMap{
						"id": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			id, _ := m["id"].(string)
			err := validate.UUID("id", id)
			if err != nil {
				return nil, err
			}
			p.Context = log.WithField(p.Context, "ScheduleRuleID", id)

			schedID, err := h.c.ScheduleRuleStore.FindScheduleID(p.Context, id)
			if err != nil {
				return newScrubber(p.Context).scrub(nil, err)
			}

			err = h.c.ScheduleRuleStore.Delete(p.Context, id)
			if err != nil {
				return newScrubber(p.Context).scrub(nil, err)
			}

			return newScrubber(p.Context).scrub(h.c.ScheduleStore.FindOne(p.Context, schedID))
		},
	}
}
