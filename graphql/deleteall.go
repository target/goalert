package graphql

import (
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"

	g "github.com/graphql-go/graphql"
)

func (h *Handler) deleteAllField() *g.Field {
	return &g.Field{
		Type: h.deleteAll,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewNonNull(g.NewInputObject(g.InputObjectConfig{
					Name: "DeleteAllInput",
					Description: "Deletes up to any number of escalation policies, steps, services, integration keys," +
						"rotations, participants, schedules, and schedule rules.",
					Fields: g.InputObjectConfigFieldMap{
						"escalation_policy_ids":      &g.InputObjectFieldConfig{Type: g.NewList(g.String)},
						"escalation_policy_step_ids": &g.InputObjectFieldConfig{Type: g.NewList(g.String)},
						"service_ids":                &g.InputObjectFieldConfig{Type: g.NewList(g.String)},
						"integration_key_ids":        &g.InputObjectFieldConfig{Type: g.NewList(g.String)},
						"rotation_ids":               &g.InputObjectFieldConfig{Type: g.NewList(g.String)},
						"rotation_participant_ids":   &g.InputObjectFieldConfig{Type: g.NewList(g.String)},
						"schedule_ids":               &g.InputObjectFieldConfig{Type: g.NewList(g.String)},
						"schedule_rule_ids":          &g.InputObjectFieldConfig{Type: g.NewList(g.String)},
						"heartbeat_monitor_ids":      &g.InputObjectFieldConfig{Type: g.NewList(g.String)},
						"user_override_ids":          &g.InputObjectFieldConfig{Type: g.NewList(g.String)},
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
			getSlice := func(s string) []string {
				v, _ := m[s].([]interface{})

				count += len(v)
				if count > limitTotal {
					v = v[:0]
					err = validate.Many(err, validation.NewFieldError(s, "too many items"))
				}
				var strs []string
				for _, i := range v {
					if str, ok := i.(string); ok {
						strs = append(strs, str)
					} else {
						err = validate.Many(err, validation.NewFieldError(s, "expected string"))
					}
				}

				return strs
			}

			var data deleteAllData
			data.EscalationPolicyIDs = getSlice("escalation_policy_ids")
			data.EscalationPolicyStepIDs = getSlice("escalation_policy_step_ids")
			data.ServiceIDs = getSlice("service_ids")
			data.IntegrationKeyIDs = getSlice("integration_key_ids")
			data.RotationIDs = getSlice("rotation_ids")
			data.RotationParticipantIDs = getSlice("rotation_participant_ids")
			data.ScheduleIDs = getSlice("schedule_ids")
			data.ScheduleRuleIDs = getSlice("schedule_rule_ids")
			data.HeartbeatMonitorIDs = getSlice("heartbeat_monitor_ids")
			data.UserOverrideIDs = getSlice("user_override_ids")

			return scrub(data, h.c.deleteAll(p.Context, &data))
		},
	}
}

func (h *Handler) deleteAllFields() g.Fields {
	return g.Fields{
		"escalation_policy_ids":      &g.Field{Type: g.NewList(g.String)},
		"escalation_policy_step_ids": &g.Field{Type: g.NewList(g.String)},
		"service_ids":                &g.Field{Type: g.NewList(g.String)},
		"integration_key_ids":        &g.Field{Type: g.NewList(g.String)},
		"rotation_ids":               &g.Field{Type: g.NewList(g.String)},
		"rotation_participant_ids":   &g.Field{Type: g.NewList(g.String)},
		"schedule_ids":               &g.Field{Type: g.NewList(g.String)},
		"schedule_rule_ids":          &g.Field{Type: g.NewList(g.String)},
		"heartbeat_monitor_ids":      &g.Field{Type: g.NewList(g.String)},
		"user_override_ids":          &g.Field{Type: g.NewList(g.String)},
	}
}
