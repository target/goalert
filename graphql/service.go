package graphql

import (
	"errors"
	"fmt"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/service"
	"github.com/target/goalert/validation"

	g "github.com/graphql-go/graphql"
)

func getService(src interface{}) (*service.Service, error) {
	switch s := src.(type) {
	case service.Service:
		return &s, nil
	case *service.Service:
		return s, nil
	default:
		return nil, fmt.Errorf("unknown source type %T", s)
	}
}
func (h *Handler) serviceOnCallUserFields() g.Fields {
	return g.Fields{
		"user_id":     &g.Field{Type: g.String},
		"user_name":   &g.Field{Type: g.String},
		"step_number": &g.Field{Type: g.Int},
	}
}
func (h *Handler) serviceFields() g.Fields {
	return g.Fields{
		"id":          &g.Field{Type: g.String},
		"name":        &g.Field{Type: g.String},
		"description": &g.Field{Type: g.String},
		"is_user_favorite": &g.Field{
			Type:        g.Boolean,
			Description: "Indicates this service has been marked as a favorite by the user.",
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				svc, err := getService(p.Source)
				if err != nil {
					return nil, err
				}
				return svc.IsUserFavorite(), nil
			},
		},
		"escalation_policy_id": &g.Field{Type: g.String},
		"escalation_policy_name": &g.Field{
			Type: g.String,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				svc, err := getService(p.Source)
				if err != nil {
					return nil, err
				}
				return svc.EscalationPolicyName(), nil
			},
		},
		"labels": &g.Field{
			Type: g.NewList(h.label),
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				svc, err := getService(p.Source)
				if err != nil {
					return nil, err
				}
				scrub := newScrubber(p.Context).scrub
				return scrub(h.c.LabelStore.FindAllByService(p.Context, svc.ID))
			},
		},
		"escalation_policy": &g.Field{
			Type: h.escalationPolicy,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				svc, err := getService(p.Source)
				if err != nil {
					return nil, err
				}
				scrub := newScrubber(p.Context).scrub
				return scrub(h.c.EscalationStore.FindOnePolicy(p.Context, svc.EscalationPolicyID))
			},
		},
		"integration_keys": &g.Field{
			Type: g.NewList(h.integrationKey),
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				s, err := getService(p.Source)
				if err != nil {
					return nil, err
				}
				return newScrubber(p.Context).scrub(h.c.IntegrationKeyStore.FindAllByService(p.Context, s.ID))
			},
		},
		"heartbeat_monitors": &g.Field{
			Type: g.NewList(h.heartbeat),
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				s, err := getService(p.Source)
				if err != nil {
					return nil, err
				}
				return newScrubber(p.Context).scrub(h.c.HeartbeatStore.FindAllByService(p.Context, s.ID))
			},
		},
		"on_call_users": &g.Field{
			Type: g.NewList(h.serviceOnCallUser),
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				s, err := getService(p.Source)
				if err != nil {
					return nil, err
				}
				return newScrubber(p.Context).scrub(h.c.OnCallStore.OnCallUsersByService(p.Context, s.ID))
			},
		},
		"alerts": &g.Field{
			Type: g.NewList(h.alert),
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				s, err := getService(p.Source)
				if err != nil {
					return nil, err
				}
				a, _, err := h.c.AlertStore.LegacySearch(p.Context, &alert.LegacySearchOptions{ServiceID: s.ID})
				return newScrubber(p.Context).scrub(a, err)
			},
		},
	}
}

func (h *Handler) serviceField() *g.Field {
	return &g.Field{
		Type: h.service,
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
			userID := permission.UserID(p.Context)

			return newScrubber(p.Context).scrub(h.c.ServiceStore.FindOneForUser(p.Context, userID, id))
		},
	}
}

func (h *Handler) servicesField() *g.Field {
	return &g.Field{
		Type:              g.NewList(h.service),
		DeprecationReason: "Use services2 instead.",
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			return newScrubber(p.Context).scrub(h.c.ServiceStore.FindAll(p.Context))
		},
	}
}

func (h *Handler) searchServicesField() *g.Field {
	return &g.Field{
		Args: g.FieldConfigArgument{
			"options": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "ServiceSearchOptions",
					Fields: g.InputObjectConfigFieldMap{
						"search": &g.InputObjectFieldConfig{
							Type:        g.String,
							Description: "Searches for case-insensitive service name or description substring match.",
						},
						"favorites_only":  &g.InputObjectFieldConfig{Description: "Only include services marked as favorites by the current user.", Type: g.Boolean},
						"favorites_first": &g.InputObjectFieldConfig{Description: "Raise favorite services to the top of results.", Type: g.Boolean},
						"limit":           &g.InputObjectFieldConfig{Description: "Limit the number of results.", Type: g.Int},
					},
				}),
			},
		},
		Type: g.NewObject(g.ObjectConfig{
			Name: "ServiceSearchResult",
			Fields: g.Fields{
				"items":       &g.Field{Type: g.NewList(h.service)},
				"total_count": &g.Field{DeprecationReason: "Preserved for compatibility, represents the length of `items`. Will never be greater than `limit`.", Type: g.Int},
			},
		}),
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			var result struct {
				Items []service.Service `json:"items"`
				Total int               `json:"total_count"`
			}
			var opts service.LegacySearchOptions
			if m, ok := p.Args["options"].(map[string]interface{}); ok {
				opts.Search, _ = m["search"].(string)
				opts.FavoritesOnly, _ = m["favorites_only"].(bool)
				opts.FavoritesFirst, _ = m["favorites_first"].(bool)
				opts.Limit, _ = m["limit"].(int)
			}
			opts.FavoritesUserID = permission.UserID(p.Context)

			s, err := h.c.ServiceStore.LegacySearch(p.Context, &opts)
			if err != nil {
				return newScrubber(p.Context).scrub(nil, err)
			}

			result.Items = s
			result.Total = len(s)
			return newScrubber(p.Context).scrub(result, err)
		},
	}
}

func (h *Handler) createServiceField() *g.Field {
	return &g.Field{
		Type: h.service,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Description: "because bugs.",
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "CreateServiceInput",
					Fields: g.InputObjectConfigFieldMap{
						"name":                 &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"description":          &g.InputObjectFieldConfig{Type: g.String},
						"escalation_policy_id": &g.InputObjectFieldConfig{Type: g.String},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			scrub := newScrubber(p.Context).scrub
			var s service.Service
			s.Name, _ = m["name"].(string)
			s.Description, _ = m["description"].(string)
			s.EscalationPolicyID, _ = m["escalation_policy_id"].(string)

			return scrub(h.c.ServiceStore.Insert(p.Context, &s))
		},
	}
}

func (h *Handler) updateServiceField() *g.Field {
	return &g.Field{
		Type: h.service,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Description: "because bugs.",
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "UpdateServiceInput",
					Fields: g.InputObjectConfigFieldMap{
						"id":                   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"name":                 &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"description":          &g.InputObjectFieldConfig{Type: g.String},
						"escalation_policy_id": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			scrub := newScrubber(p.Context).scrub

			var s service.Service
			s.ID, _ = m["id"].(string)
			s.Name, _ = m["name"].(string)
			s.Description, _ = m["description"].(string)
			s.EscalationPolicyID, _ = m["escalation_policy_id"].(string)
			err := h.c.ServiceStore.Update(p.Context, &s)
			if err != nil {
				return scrub(nil, err)
			}
			userID := permission.UserID(p.Context)
			return scrub(h.c.ServiceStore.FindOneForUser(p.Context, userID, s.ID))
		},
	}
}

func (h *Handler) deleteServiceField() *g.Field {
	return &g.Field{
		Type: g.NewObject(g.ObjectConfig{
			Name:   "DeleteServiceOutput",
			Fields: g.Fields{"deleted_service_id": &g.Field{Type: g.String}},
		}),
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Description: "because bugs.",
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "DeleteServiceInput",
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

			var s struct {
				ID string `json:"deleted_service_id"`
			}

			s.ID, _ = m["id"].(string)

			err := h.c.ServiceStore.Delete(p.Context, s.ID)
			return newScrubber(p.Context).scrub(&s, err)
		},
	}
}
