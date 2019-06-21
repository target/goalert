package graphql

import (
	"errors"
	"github.com/target/goalert/user/contactmethod"

	g "github.com/graphql-go/graphql"
)

func (h *Handler) CMFields() g.Fields {
	return g.Fields{
		"id":       &g.Field{Type: g.String},
		"name":     &g.Field{Type: g.String},
		"type":     &g.Field{Type: g.String},
		"value":    &g.Field{Type: g.String},
		"disabled": &g.Field{Type: g.Boolean},
	}
}

var contactType = g.NewEnum(g.EnumConfig{
	Name: "ContactType",
	Values: g.EnumValueConfigMap{
		"VOICE": &g.EnumValueConfig{Value: contactmethod.TypeVoice},
		"SMS":   &g.EnumValueConfig{Value: contactmethod.TypeSMS},
		"EMAIL": &g.EnumValueConfig{Value: contactmethod.TypeEmail},
		"PUSH":  &g.EnumValueConfig{Value: contactmethod.TypePush},
	},
})

func (h *Handler) updateContactMethodField() *g.Field {
	return &g.Field{
		Type: h.contactMethod,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Description: "because bugs.",
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "UpdateContactMethodInput",
					Fields: g.InputObjectConfigFieldMap{
						"id":       &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"name":     &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"type":     &g.InputObjectFieldConfig{Type: g.NewNonNull(contactType)},
						"value":    &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"disabled": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Boolean)},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			var c contactmethod.ContactMethod
			c.ID, _ = m["id"].(string)
			c.Name, _ = m["name"].(string)
			c.Type, _ = m["type"].(contactmethod.Type)
			c.Value, _ = m["value"].(string)
			c.Disabled, _ = m["disabled"].(bool)

			err := h.c.CMStore.Update(p.Context, &c)
			if err != nil {
				return newScrubber(p.Context).scrub(nil, err)
			}
			return newScrubber(p.Context).scrub(h.c.CMStore.FindOne(p.Context, c.ID))
		},
	}
}

func (h *Handler) createContactMethodField() *g.Field {
	return &g.Field{
		Type: h.contactMethod,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Description: "because bugs.",
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "CreateContactMethodInput",
					Fields: g.InputObjectConfigFieldMap{
						"name":     &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"type":     &g.InputObjectFieldConfig{Type: g.NewNonNull(contactType)},
						"value":    &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"disabled": &g.InputObjectFieldConfig{Type: g.Boolean},
						"user_id":  &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			var c contactmethod.ContactMethod
			c.Name, _ = m["name"].(string)
			c.Type, _ = m["type"].(contactmethod.Type)
			c.Value, _ = m["value"].(string)
			c.Disabled, _ = m["disabled"].(bool)
			c.UserID, _ = m["user_id"].(string)

			return newScrubber(p.Context).scrub(h.c.CMStore.Insert(p.Context, &c))
		},
	}
}

func (h *Handler) deleteContactMethodField() *g.Field {
	return &g.Field{
		Type: g.NewObject(g.ObjectConfig{
			Name:   "DeleteContactMethodOutput",
			Fields: g.Fields{"deleted_id": &g.Field{Type: g.String}},
		}),
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Description: "because bugs.",
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "DeleteContactMethodInput",
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

			var c struct {
				ID string `json:"deleted_id"`
			}

			c.ID, _ = m["id"].(string)

			err := h.c.CMStore.Delete(p.Context, c.ID)
			return newScrubber(p.Context).scrub(&c, err)
		},
	}
}

func (h *Handler) sendContactMethodTest() *g.Field {
	return &g.Field{
		Type: g.NewObject(g.ObjectConfig{
			Name:   "SendContactMethodTest",
			Fields: g.Fields{"id": &g.Field{Type: g.String}},
		}),
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "SendContactMethodTestInput",
					Fields: g.InputObjectConfigFieldMap{
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

			var result struct {
				CMID string `json:"id"`
			}
			result.CMID, _ = m["contact_method_id"].(string)

			err := h.c.NotificationStore.SendContactMethodTest(p.Context, result.CMID)
			return newScrubber(p.Context).scrub(result, err)
		},
	}
}
