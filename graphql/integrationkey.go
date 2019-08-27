package graphql

import (
	"errors"
	"github.com/target/goalert/config"
	"github.com/target/goalert/validation"
	"net/url"

	"github.com/target/goalert/integrationkey"

	"fmt"

	g "github.com/graphql-go/graphql"
)

func (h *Handler) integrationKeyFields() g.Fields {
	return g.Fields{
		"id":         &g.Field{Type: g.String},
		"name":       &g.Field{Type: g.String},
		"type":       &g.Field{Type: integrationKeyType},
		"service_id": &g.Field{Type: g.String},
		"service": &g.Field{
			Type: h.service,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				var ID string
				switch i := p.Source.(type) {
				case integrationkey.IntegrationKey:
					ID = i.ServiceID
				case *integrationkey.IntegrationKey:
					ID = i.ServiceID
				default:
					return nil, fmt.Errorf("could not resolve ServiceID of integration_key (unknown source type %T)", i)
				}

				return newScrubber(p.Context).scrub(h.c.ServiceStore.FindOne(p.Context, ID))
			},
		},
		"href": &g.Field{
			Type: g.String,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				cfg := config.FromContext(p.Context)
				var key integrationkey.IntegrationKey
				switch i := p.Source.(type) {
				case integrationkey.IntegrationKey:
					key = i
				case *integrationkey.IntegrationKey:
					key = *i
				default:
					return nil, fmt.Errorf("error resolving key ID (unknown source type %T)", i)
				}

				switch key.Type {
				case integrationkey.TypeGeneric:
					return "/v1/api/alerts?integration_key=" + url.QueryEscape(key.ID), nil
				case integrationkey.TypeGrafana:
					return "/v1/webhooks/grafana?integration_key=" + url.QueryEscape(key.ID), nil
				case integrationkey.TypeEmail:
					if !cfg.Mailgun.Enable || cfg.Mailgun.EmailDomain == "" {
						return "", nil
					}
					return "mailto:" + key.ID + "@" + cfg.Mailgun.EmailDomain, nil
				}

				return "#" + url.QueryEscape(key.ID), nil
			},
		},
	}
}

func (h *Handler) integrationKeyField() *g.Field {
	return &g.Field{
		Type: h.integrationKey,
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

			return newScrubber(p.Context).scrub(h.c.IntegrationKeyStore.FindOne(p.Context, id))
		},
	}
}

func (h *Handler) integrationKeysField() *g.Field {
	return &g.Field{
		Name: "IntegrationKeys",
		Type: g.NewList(h.integrationKey),
		Args: g.FieldConfigArgument{
			"service_id": &g.ArgumentConfig{
				Type: g.NewNonNull(g.String),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			id, ok := p.Args["service_id"].(string)
			if !ok {
				return nil, validation.NewFieldError("service_id", "required")
			}
			return newScrubber(p.Context).scrub(h.c.IntegrationKeyStore.FindAllByService(p.Context, id))
		},
	}
}

var integrationKeyType = g.NewEnum(g.EnumConfig{
	Name: "IntegrationKeyType",
	Values: g.EnumValueConfigMap{
		"grafana": &g.EnumValueConfig{Value: integrationkey.TypeGrafana},
		"generic": &g.EnumValueConfig{Value: integrationkey.TypeGeneric},
		"email":   &g.EnumValueConfig{Value: integrationkey.TypeEmail},
	},
})

func (h *Handler) createIntegrationKeyField() *g.Field {
	return &g.Field{
		Type: h.integrationKey,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Description: "because bugs.",
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "CreateIntegrationKeyInput",
					Fields: g.InputObjectConfigFieldMap{
						"name":       &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"type":       &g.InputObjectFieldConfig{Type: g.NewNonNull(integrationKeyType)},
						"service_id": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			var i integrationkey.IntegrationKey
			i.Name, _ = m["name"].(string)
			i.Type, _ = m["type"].(integrationkey.Type)
			i.ServiceID, _ = m["service_id"].(string)

			return newScrubber(p.Context).scrub(h.c.IntegrationKeyStore.Create(p.Context, &i))
		},
	}
}

func (h *Handler) deleteIntegrationKeyField() *g.Field {
	return &g.Field{
		Type: g.NewObject(g.ObjectConfig{
			Name:   "DeleteIntegrationKeyOutput",
			Fields: g.Fields{"deleted_id": &g.Field{Type: g.String}},
		}),
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Description: "because bugs.",
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "DeleteIntegrationKeyInput",
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
				ID string `json:"deleted_id"`
			}

			r.ID, _ = m["id"].(string)

			return newScrubber(p.Context).scrub(r, h.c.IntegrationKeyStore.Delete(p.Context, r.ID))
		},
	}
}
