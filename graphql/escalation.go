package graphql

import (
	"fmt"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/user"
	"github.com/target/goalert/validation"

	g "github.com/graphql-go/graphql"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

func getPolicy(src interface{}) (*escalation.Policy, error) {
	switch t := src.(type) {
	case escalation.Policy:
		return &t, nil
	case *escalation.Policy:
		return t, nil
	default:
		return nil, fmt.Errorf("invalid source type for escalation policy %T", t)
	}
}

func getStep(src interface{}) (*escalation.Step, error) {
	switch t := src.(type) {
	case escalation.Step:
		return &t, nil
	case *escalation.Step:
		return t, nil
	default:
		return nil, fmt.Errorf("invalid source type for escalation step %T", t)
	}
}

func (h *Handler) escalationPolicyField() *g.Field {
	return &g.Field{
		Type: h.escalationPolicy,
		Args: g.FieldConfigArgument{
			"id": &g.ArgumentConfig{
				Type: g.NewNonNull(g.String),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			id, ok := p.Args["id"].(string)
			if !ok {
				return nil, validation.NewFieldError("id", "required")
			}

			return newScrubber(p.Context).scrub(h.c.EscalationStore.FindOnePolicy(p.Context, id))
		},
	}
}

func (h *Handler) escalationPoliciesField() *g.Field {
	return &g.Field{
		Name: "EscalationPolicies",
		Type: g.NewList(h.escalationPolicy),
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			return newScrubber(p.Context).scrub(h.c.EscalationStore.FindAllPolicies(p.Context))
		},
	}
}

func (h *Handler) createOrUpdateEscalationPolicyField() *g.Field {
	return &g.Field{
		Type: h.escalationPolicy,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "CreateOrUpdateEscalationPolicyInput",
					Fields: g.InputObjectConfigFieldMap{
						"id":          &g.InputObjectFieldConfig{Type: g.String},
						"name":        &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"description": &g.InputObjectFieldConfig{Type: g.String},
						"repeat":      &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Int)},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input")
			}

			ep := new(escalation.Policy)
			ep.ID, _ = m["id"].(string)
			ep.Name, _ = m["name"].(string)
			ep.Description, _ = m["description"].(string)
			ep.Repeat, _ = m["repeat"].(int)

			if ep.ID == "" {
				return newScrubber(p.Context).scrub(h.c.EscalationStore.CreatePolicy(p.Context, ep))
			}

			return newScrubber(p.Context).scrub(ep, h.c.EscalationStore.UpdatePolicy(p.Context, ep))
		},
	}
}

func (h *Handler) deleteEscalationPolicyField() *g.Field {
	return &g.Field{
		Type: g.NewObject(g.ObjectConfig{
			Name:   "DeleteEscalationPolicyOutput",
			Fields: g.Fields{"deleted_id": &g.Field{Type: g.String}},
		}),
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "DeleteEscalationPolicyInput",
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

			var n struct {
				ID string `json:"deleted_id"`
			}

			n.ID, _ = m["id"].(string)

			err := h.c.EscalationStore.DeletePolicy(p.Context, n.ID)

			// Code 23503 corresponds to: "foreign_key_violation"
			// https://www.postgresql.org/docs/9.6/static/errcodes-appendix.html
			if e, ok := errors.Cause(err).(*pq.Error); ok && e.Code == "23503" && e.Constraint == "services_escalation_policy_id_fkey" {
				return nil, errors.New("policy is currently in use by one or more services")
			}
			return newScrubber(p.Context).scrub(&n, err)
		},
	}
}

func (h *Handler) escalationPolicyFields() g.Fields {
	return g.Fields{
		"id":          &g.Field{Type: g.String},
		"name":        &g.Field{Type: g.String},
		"description": &g.Field{Type: g.String},
		"repeat":      &g.Field{Type: g.Int},
		"target_type": targetTypeField(assignment.TargetTypeEscalationPolicy),
		"services": &g.Field{
			Type:        g.NewList(h.service),
			Description: "List of services currently assigned to this escalation policy.",
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				ep, err := getPolicy(p.Source)
				if err != nil {
					return nil, err
				}
				scrub := newScrubber(p.Context).scrub

				return scrub(h.c.ServiceStore.FindAllByEP(p.Context, ep.ID))
			},
		},

		"steps": &g.Field{
			Type: g.NewList(h.escalationPolicyStep),
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				ep, err := getPolicy(p.Source)
				if err != nil {
					return nil, err
				}

				return newScrubber(p.Context).scrub(h.c.EscalationStore.FindAllSteps(p.Context, ep.ID))
			},
		},
	}
}

func (h *Handler) escalationPolicyStepFields() g.Fields {
	return g.Fields{
		"id":                   &g.Field{Type: g.String},
		"escalation_policy_id": &g.Field{Type: g.String},
		"delay_minutes":        &g.Field{Type: g.Int},
		"step_number":          &g.Field{Type: g.Int},
		"user_ids": &g.Field{
			Type:              g.NewList(g.String),
			DeprecationReason: "Use the 'targets' field instead.",
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				s, err := getStep(p.Source)
				if err != nil {
					return nil, err
				}
				tgts, err := h.c.EscalationStore.FindAllStepTargets(p.Context, s.ID)
				if err != nil {
					return newScrubber(p.Context).scrub(nil, err)
				}
				result := make([]string, 0, len(tgts))
				for _, t := range tgts {
					if t.TargetType() != assignment.TargetTypeUser {
						continue
					}
					result = append(result, t.TargetID())
				}
				return result, err
			},
		},
		"schedule_ids": &g.Field{
			Type:              g.NewList(g.String),
			DeprecationReason: "Use the 'targets' field instead.",
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				s, err := getStep(p.Source)
				if err != nil {
					return nil, err
				}
				tgts, err := h.c.EscalationStore.FindAllStepTargets(p.Context, s.ID)
				if err != nil {
					return newScrubber(p.Context).scrub(nil, err)
				}

				result := make([]string, 0, len(tgts))
				for _, t := range tgts {
					if t.TargetType() != assignment.TargetTypeSchedule {
						continue
					}
					result = append(result, t.TargetID())
				}
				return result, err
			},
		},
		"source_type": sourceTypeField(assignment.SrcTypeEscalationPolicyStep),
		"source":      h.sourceField(),

		"escalation_policy": &g.Field{
			Type: h.escalationPolicy,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				s, err := getStep(p.Source)
				if err != nil {
					return nil, err
				}

				return newScrubber(p.Context).scrub(h.c.EscalationStore.FindOnePolicy(p.Context, s.PolicyID))
			},
		},

		"users": &g.Field{
			Type:              g.NewList(h.user),
			DeprecationReason: "Use the 'targets' field instead.",
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				s, err := getStep(p.Source)
				if err != nil {
					return nil, err
				}
				var result []user.User
				var u *user.User
				tgts, err := h.c.EscalationStore.FindAllStepTargets(p.Context, s.ID)
				if err != nil {
					return newScrubber(p.Context).scrub(nil, err)
				}

				for _, t := range tgts {
					if t.TargetType() != assignment.TargetTypeUser {
						continue
					}
					u, err = h.c.UserStore.FindOne(p.Context, t.TargetID())
					if err != nil {
						return newScrubber(p.Context).scrub(nil, err)
					}
					result = append(result, *u)
				}

				return result, nil
			},
		},

		"targets": &g.Field{
			Type: g.NewList(h.assignmentTarget),
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				s, err := getStep(p.Source)
				if err != nil {
					return nil, err
				}
				tgts, err := h.c.EscalationStore.FindAllStepTargets(p.Context, s.ID)
				if err != nil {
					return newScrubber(p.Context).scrub(nil, err)
				}

				return tgts, nil
			},
		},

		"schedules": &g.Field{
			Type:              g.NewList(h.schedule),
			DeprecationReason: "Use the 'targets' field instead.",
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				s, err := getStep(p.Source)
				if err != nil {
					return nil, err
				}
				tgts, err := h.c.EscalationStore.FindAllStepTargets(p.Context, s.ID)
				if err != nil {
					return newScrubber(p.Context).scrub(nil, err)
				}

				var result []schedule.Schedule
				var u *schedule.Schedule
				for _, t := range tgts {
					if t.TargetType() != assignment.TargetTypeSchedule {
						continue
					}
					u, err = h.c.ScheduleStore.FindOne(p.Context, t.TargetID())
					if err != nil {
						return newScrubber(p.Context).scrub(nil, err)
					}
					result = append(result, *u)
				}

				return result, nil
			},
		},
	}
}

var epStepTarget = g.NewEnum(g.EnumConfig{
	Name: "EscalationPolicyStepTarget",
	Values: g.EnumValueConfigMap{
		"user":     &g.EnumValueConfig{Value: assignment.TargetTypeUser},
		"schedule": &g.EnumValueConfig{Value: assignment.TargetTypeSchedule},
		"rotation": &g.EnumValueConfig{Value: assignment.TargetTypeRotation},
	},
})

func (h *Handler) addEscalationPolicyStepTargetField() *g.Field {
	return &g.Field{
		Type: h.assignmentTarget,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "AddEscalationPolicyStepTargetInput",
					Fields: g.InputObjectConfigFieldMap{
						"step_id":     &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"target_id":   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"target_type": &g.InputObjectFieldConfig{Type: g.NewNonNull(epStepTarget)},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}
			var tgt assignment.RawTarget
			tgt.ID, _ = m["target_id"].(string)
			tgt.Type, _ = m["target_type"].(assignment.TargetType)
			stepID, _ := m["step_id"].(string)

			err := h.c.EscalationStore.AddStepTarget(p.Context, stepID, tgt)
			if err != nil {
				return newScrubber(p.Context).scrub(nil, err)
			}

			return tgt, nil
		},
	}
}

func (h *Handler) deleteEscalationPolicyStepTargetField() *g.Field {
	return &g.Field{
		Type: g.NewObject(g.ObjectConfig{
			Name: "DeleteEscalationPolicyStepTargetOutput",
			Fields: g.Fields{
				"target_id":   &g.Field{Type: g.String, Description: "ID of the target."},
				"target_type": &g.Field{Type: epStepTarget, Description: "The type of the target."},
			},
		}),
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "DeleteEscalationPolicyStepTargetInput",
					Fields: g.InputObjectConfigFieldMap{
						"step_id":     &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"target_id":   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"target_type": &g.InputObjectFieldConfig{Type: g.NewNonNull(epStepTarget)},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}
			var tgt assignment.RawTarget
			tgt.ID, _ = m["target_id"].(string)
			tgt.Type, _ = m["target_type"].(assignment.TargetType)
			stepID, _ := m["step_id"].(string)
			err := h.c.EscalationStore.DeleteStepTarget(p.Context, stepID, tgt)
			return newScrubber(p.Context).scrub(tgt, err)
		},
	}
}

func (h *Handler) createOrUpdateEscalationPolicyStepField() *g.Field {
	return &g.Field{
		Type: g.NewObject(g.ObjectConfig{
			Name: "CreateOrUpdateEscalationPolicyStepOutput",
			Fields: g.Fields{
				"created":                &g.Field{Type: g.Boolean, Description: "Signifies if a new record was created."},
				"escalation_policy_step": &g.Field{Type: h.escalationPolicyStep, Description: "The created or updated record."},
			},
		}),
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "CreateOrUpdateEscalationPolicyStepInput",
					Fields: g.InputObjectConfigFieldMap{
						"id":                   &g.InputObjectFieldConfig{Type: g.String},
						"escalation_policy_id": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"delay_minutes":        &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Int)},
						"user_ids":             &g.InputObjectFieldConfig{Type: g.NewNonNull(g.NewList(g.String))},
						"schedule_ids":         &g.InputObjectFieldConfig{Type: g.NewNonNull(g.NewList(g.String))},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}
			var r struct {
				Created bool             `json:"created"`
				S       *escalation.Step `json:"escalation_policy_step"`
			}
			r.S = new(escalation.Step)
			var hasID bool
			r.S.ID, hasID = m["id"].(string)
			r.S.PolicyID, _ = m["escalation_policy_id"].(string)
			r.S.DelayMinutes, _ = m["delay_minutes"].(int)

			userIDs, _ := m["user_ids"].([]interface{})
			schedIDs, _ := m["schedule_ids"].([]interface{})

			asn := make([]assignment.Target, 0, len(userIDs)+len(schedIDs))
			for _, _id := range userIDs {
				if id, ok := _id.(string); ok {
					asn = append(asn, assignment.UserTarget(id))
				}
			}

			for _, _id := range schedIDs {
				if id, ok := _id.(string); ok {
					asn = append(asn, assignment.ScheduleTarget(id))
				}
			}

			scrub := newScrubber(p.Context).scrub
			var err error
			if !hasID {
				r.S, err = h.c.EscalationStore.CreateStep(p.Context, r.S)
				r.Created = true
				if err != nil {
					return newScrubber(p.Context).scrub(nil, err)
				}
				for _, tgt := range asn {
					err = h.c.EscalationStore.AddStepTarget(p.Context, r.S.ID, tgt)
					if err != nil {
						return scrub(nil, err)
					}
				}
				return r, nil
			}

			err = h.c.EscalationStore.UpdateStep(p.Context, r.S)
			if err != nil {
				return scrub(nil, err)
			}
			old, err := h.c.EscalationStore.FindAllStepTargets(p.Context, r.S.ID)
			if err != nil {
				return scrub(nil, err)
			}

			added := make(map[assignment.RawTarget]bool, len(asn))
			for _, tgt := range asn {
				added[assignment.RawTarget{Type: tgt.TargetType(), ID: tgt.TargetID()}] = true
				err = h.c.EscalationStore.AddStepTarget(p.Context, r.S.ID, tgt)
				if err != nil {
					return scrub(nil, err)
				}
			}

			for _, tgt := range old {
				if added[assignment.RawTarget{Type: tgt.TargetType(), ID: tgt.TargetID()}] {
					continue
				}
				if tgt.TargetType() == assignment.TargetTypeRotation {
					continue
				}

				err = h.c.EscalationStore.DeleteStepTarget(p.Context, r.S.ID, tgt)
				if err != nil {
					return scrub(nil, err)
				}
			}

			r.S, err = h.c.EscalationStore.FindOneStep(p.Context, r.S.ID)
			return scrub(r, err)
		},
	}
}

func (h *Handler) deleteEscalationPolicyStepField() *g.Field {
	return &g.Field{
		Description: "Remove a step from an escalation policy.",
		Type: g.NewObject(g.ObjectConfig{
			Name: "DeleteEscalationPolicyStepOutput",
			Fields: g.Fields{
				"deleted_id":           &g.Field{Type: g.String},
				"escalation_policy_id": &g.Field{Type: g.String},
			},
		}),
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "DeleteEscalationPolicyStepInput",
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

			var r struct {
				ID   string `json:"deleted_id"`
				EPID string `json:"escalation_policy_id"`
			}
			r.ID, _ = m["id"].(string)
			var err error
			r.EPID, err = h.c.EscalationStore.DeleteStep(p.Context, r.ID)
			return newScrubber(p.Context).scrub(r, err)
		},
	}
}

func (h *Handler) moveEscalationPolicyStepField() *g.Field {
	return &g.Field{
		Description: "Moves a step to new_position, automatically shifting other participants around.",
		Type:        g.NewList(h.escalationPolicyStep),
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "MoveEscalationPolicyStepInput",
					Fields: g.InputObjectConfigFieldMap{
						"id":           &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"new_position": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Int)},
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
			newPos, _ := m["new_position"].(int)

			err := h.c.EscalationStore.MoveStep(p.Context, id, newPos)
			if err != nil {
				return newScrubber(p.Context).scrub(nil, err)
			}

			eps, err := h.c.EscalationStore.FindOneStep(p.Context, id)
			if err != nil {
				return newScrubber(p.Context).scrub(nil, err)
			}
			return newScrubber(p.Context).scrub(h.c.EscalationStore.FindAllSteps(p.Context, eps.PolicyID))
		},
	}
}
