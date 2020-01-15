package graphql

import (
	"errors"
	"github.com/target/goalert/limit"

	g "github.com/graphql-go/graphql"
)

var limitID = g.NewEnum(g.EnumConfig{
	Name: "LimitID",
	Values: g.EnumValueConfigMap{
		"notification_rules_per_user":    &g.EnumValueConfig{Value: limit.NotificationRulesPerUser},
		"contact_methods_per_user":       &g.EnumValueConfig{Value: limit.ContactMethodsPerUser},
		"ep_steps_per_policy":            &g.EnumValueConfig{Value: limit.EPStepsPerPolicy},
		"ep_actions_per_step":            &g.EnumValueConfig{Value: limit.EPActionsPerStep},
		"participants_per_rotation":      &g.EnumValueConfig{Value: limit.ParticipantsPerRotation},
		"rules_per_schedule":             &g.EnumValueConfig{Value: limit.RulesPerSchedule},
		"integration_keys_per_service":   &g.EnumValueConfig{Value: limit.IntegrationKeysPerService},
		"unacked_alerts_per_service":     &g.EnumValueConfig{Value: limit.UnackedAlertsPerService},
		"targets_per_schedule":           &g.EnumValueConfig{Value: limit.TargetsPerSchedule},
		"heartbeat_monitors_per_service": &g.EnumValueConfig{Value: limit.HeartbeatMonitorsPerService},
		"user_overrides_per_schedule":    &g.EnumValueConfig{Value: limit.UserOverridesPerSchedule},
	},
})

func (h *Handler) updateConfigLimitField() *g.Field {
	return &g.Field{
		Type: g.NewObject(g.ObjectConfig{
			Name: "UpdateConfigLimitOutput",
			Fields: g.Fields{
				"id":  &g.Field{Type: limitID},
				"max": &g.Field{Type: g.Int},
			},
		}),
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewNonNull(g.NewInputObject(g.InputObjectConfig{
					Name: "UpdateConfigLimitInput",
					Fields: g.InputObjectConfigFieldMap{
						"id":  &g.InputObjectFieldConfig{Type: g.NewNonNull(limitID)},
						"max": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Int)},
					},
				})),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			var res struct {
				ID  limit.ID `json:"id"`
				Max int      `json:"max"`
			}

			res.ID, _ = m["id"].(limit.ID)
			res.Max, _ = m["max"].(int)

			err := h.c.LimitStore.SetMax(p.Context, res.ID, res.Max)
			return newScrubber(p.Context).scrub(res, err)
		},
	}
}
