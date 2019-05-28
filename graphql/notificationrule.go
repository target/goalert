package graphql

import (
	"errors"
	"github.com/target/goalert/user/notificationrule"

	"fmt"

	g "github.com/graphql-go/graphql"
)

func (h *Handler) NRFields() g.Fields {
	return g.Fields{
		"id":            &g.Field{Type: g.String},
		"delay_minutes": &g.Field{Type: g.Int, Description: "Delay in minutes."},
		"delay": &g.Field{
			Type:              g.Int,
			DeprecationReason: "use 'delay_minutes' instead",
			Resolve: func(p g.ResolveParams) (interface{}, error) {

				var n notificationrule.NotificationRule

				switch t := p.Source.(type) {
				case notificationrule.NotificationRule:
					n = t
				case *notificationrule.NotificationRule:
					n = *t
				default:
					return nil, fmt.Errorf("invalid source type for notification rule %T", t)
				}

				return n.DelayMinutes, nil
			},
		},
		"contact_method_id": &g.Field{Type: g.String},
		"contact_method": &g.Field{
			Type: h.contactMethod,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				var n notificationrule.NotificationRule

				switch t := p.Source.(type) {
				case notificationrule.NotificationRule:
					n = t
				case *notificationrule.NotificationRule:
					n = *t
				default:
					return nil, fmt.Errorf("invalid source type for notification rule %T", t)
				}

				return newScrubber(p.Context).scrub(h.c.CMStore.FindOne(p.Context, n.ContactMethodID))
			},
		},
	}
}

func (h *Handler) updateNotificationRuleField() *g.Field {
	return &g.Field{
		Type: h.notificationRule,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Description: "because bugs.",
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "UpdateNotificationRuleInput",
					Fields: g.InputObjectConfigFieldMap{
						"id":            &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"delay_minutes": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Int)},
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
			delay, _ := m["delay_minutes"].(int)

			err := h.c.NRStore.UpdateDelay(p.Context, id, delay)
			if err != nil {
				return newScrubber(p.Context).scrub(nil, err)
			}
			return newScrubber(p.Context).scrub(h.c.NRStore.FindOne(p.Context, id))
		},
	}
}

func (h *Handler) createNotificationRuleField() *g.Field {
	return &g.Field{
		Type: h.notificationRule,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Description: "because bugs.",
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "CreateNotificationRuleInput",
					Fields: g.InputObjectConfigFieldMap{
						"user_id":           &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"delay_minutes":     &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Int)},
						"contact_method_id": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			var n notificationrule.NotificationRule

			n.UserID, _ = m["user_id"].(string)
			n.DelayMinutes, _ = m["delay_minutes"].(int)
			n.ContactMethodID, _ = m["contact_method_id"].(string)

			return newScrubber(p.Context).scrub(h.c.NRStore.Insert(p.Context, &n))
		},
	}
}

func (h *Handler) deleteNotificationRuleField() *g.Field {
	return &g.Field{
		Type: g.NewObject(g.ObjectConfig{
			Name:   "DeleteNotificationRuleOutput",
			Fields: g.Fields{"deleted_id": &g.Field{Type: g.String}},
		}),
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Description: "because bugs.",
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "DeleteNotificationRuleInput",
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

			err := h.c.NRStore.Delete(p.Context, n.ID)
			return newScrubber(p.Context).scrub(&n, err)
		},
	}
}
