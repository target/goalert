package graphql

import (
	g "github.com/graphql-go/graphql"
)

func (h *Handler) onCallAssignmentFields() g.Fields {
	return g.Fields{
		"is_active":                     &g.Field{Type: g.Boolean},
		"service_id":                    &g.Field{Type: g.String},
		"service_name":                  &g.Field{Type: g.String},
		"escalation_policy_id":          &g.Field{Type: g.String},
		"escalation_policy_name":        &g.Field{Type: g.String},
		"escalation_policy_step_number": &g.Field{Type: g.Int},
		"rotation_id":                   &g.Field{Type: g.String},
		"rotation_name":                 &g.Field{Type: g.String},
		"schedule_id":                   &g.Field{Type: g.String},
		"schedule_name":                 &g.Field{Type: g.String},
		"user_id":                       &g.Field{Type: g.String},
	}
}
