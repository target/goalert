package graphql

import (
	"errors"
	"fmt"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/user"
	"github.com/target/goalert/validation"

	g "github.com/graphql-go/graphql"
)

var userRoleEnum = g.NewEnum(g.EnumConfig{
	Name: "UserRole",
	Values: g.EnumValueConfigMap{
		"admin": &g.EnumValueConfig{Value: permission.RoleAdmin},
		"user":  &g.EnumValueConfig{Value: permission.RoleUser},
	},
})

func getUser(src interface{}) (*user.User, error) {
	switch u := src.(type) {
	case *user.User:
		return u, nil
	case user.User:
		return &u, nil
	default:
		return nil, fmt.Errorf("could not id of user (unknown source type %T)", u)
	}
}

func (h *Handler) userFields() g.Fields {
	return g.Fields{
		"id":    &g.Field{Type: g.String},
		"name":  &g.Field{Type: g.String},
		"bio":   &g.Field{Type: g.String, DeprecationReason: "Bio is no longer used or listed on the user details page"},
		"email": &g.Field{Type: g.String},
		"avatar_url": &g.Field{
			Type:              g.String,
			DeprecationReason: "Use /v1/api/users/{userID}/avatar instead.",
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				u, err := getUser(p.Source)
				if err != nil {
					return nil, err
				}

				// Return the same URL used previously (large). We can't return the
				// new redirect-API URL here because it requires auth, and old UI code
				// using it won't provide the token, which would result in broken images.
				//
				// So, for now, this field will use the new method and effectively return
				// the same URL. Once the UI is updated, we can either remove this field
				// or point it to the redirect URL.
				return u.ResolveAvatarURL(true), nil
			},
		},
		"role": &g.Field{Type: userRoleEnum},
		"first_name": &g.Field{
			Type:              g.String,
			DeprecationReason: "use 'name' instead",
			Resolve:           func(p g.ResolveParams) (interface{}, error) { return p.Source.(*user.User).Name, nil },
		},
		"last_name": &g.Field{
			Type:              g.String,
			DeprecationReason: "use 'name' instead",
			Resolve:           func(g.ResolveParams) (interface{}, error) { return "", nil },
		},
		"alert_status_log_contact_method_id": &g.Field{Type: g.String, Description: "Configures a contact method ID to be used for automatic status updates of alerts."},

		"contact_methods": &g.Field{
			Type: g.NewList(h.contactMethod),
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				u, err := getUser(p.Source)
				if err != nil {
					return nil, err
				}

				return newScrubber(p.Context).scrub(h.c.CMStore.FindAll(p.Context, u.ID))
			},
		},

		"notification_rules": &g.Field{
			Type: g.NewList(h.notificationRule),
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				u, err := getUser(p.Source)
				if err != nil {
					return nil, err
				}

				return newScrubber(p.Context).scrub(h.c.NRStore.FindAll(p.Context, u.ID))
			},
		},

		"on_call": &g.Field{
			Type: g.Boolean,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				u, err := getUser(p.Source)
				if err != nil {
					return nil, err
				}
				scrub := newScrubber(p.Context).scrub
				return scrub(h.c.Resolver.IsUserOnCall(p.Context, u.ID))
			},
		},

		"on_call_assignments": &g.Field{
			Type: g.NewList(h.onCallAssignment),
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				u, err := getUser(p.Source)
				if err != nil {
					return nil, err
				}
				scrub := newScrubber(p.Context).scrub

				return scrub(h.c.Resolver.OnCallByUser(p.Context, u.ID))
			},
		},
	}
}

func (h *Handler) updateUserField() *g.Field {
	return &g.Field{
		Type: h.user,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Description: "Update a user with provided fields.",
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "UpdateUserInput",
					Fields: g.InputObjectConfigFieldMap{
						"id":                                 &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"name":                               &g.InputObjectFieldConfig{Type: g.String},
						"email":                              &g.InputObjectFieldConfig{Type: g.String},
						"avatar_url":                         &g.InputObjectFieldConfig{Type: g.String},
						"role":                               &g.InputObjectFieldConfig{Type: userRoleEnum},
						"alert_status_log_contact_method_id": &g.InputObjectFieldConfig{Type: g.String},
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
			tx, err := h.legacyDB.db.BeginTx(p.Context, nil)
			if err != nil {
				return scrub(nil, err)
			}
			defer tx.Rollback()

			id, _ := m["id"].(string)
			usr, err := h.c.UserStore.FindOneTx(p.Context, tx, id, true)
			if err != nil {
				return scrub(nil, err)
			}

			usr.AlertStatusCMID, _ = m["alert_status_log_contact_method_id"].(string)
			err = h.c.UserStore.UpdateTx(p.Context, tx, usr)
			if err != nil {
				return scrub(nil, err)
			}

			err = tx.Commit()
			if err != nil {
				return scrub(nil, err)
			}
			return scrub(usr, nil)
		},
	}
}

func (h *Handler) currentUserField() *g.Field {
	return &g.Field{
		Type: h.user,
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			id := permission.UserID(p.Context)
			if id == "" {
				return nil, nil
			}

			return newScrubber(p.Context).scrub(h.c.UserStore.FindOne(p.Context, id))
		},
	}
}

func (h *Handler) userField() *g.Field {
	return &g.Field{
		Type: h.user,
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

			return newScrubber(p.Context).scrub(h.c.UserStore.FindOne(p.Context, id))
		},
	}
}
func (h *Handler) usersField() *g.Field {
	return &g.Field{
		Name: "Users",
		Type: g.NewList(h.user),
		Resolve: func(p g.ResolveParams) (interface{}, error) {

			return newScrubber(p.Context).scrub(h.c.UserStore.FindAll(p.Context))
		},
	}
}
