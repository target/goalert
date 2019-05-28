package graphql

import (
	"fmt"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/override"
	"github.com/target/goalert/validation"
	"time"

	g "github.com/graphql-go/graphql"
	"github.com/pkg/errors"
)

var userOverrideTargetType = g.NewEnum(g.EnumConfig{
	Name: "UserOverrideTargetType",
	Values: g.EnumValueConfigMap{
		"schedule": &g.EnumValueConfig{Value: assignment.TargetTypeSchedule},
	},
})

func getUserOverride(src interface{}) (*override.UserOverride, error) {
	switch u := src.(type) {
	case *override.UserOverride:
		return u, nil
	case override.UserOverride:
		return &u, nil
	default:
		return nil, fmt.Errorf("could not get UserOverride (unknown source type %T)", u)
	}
}

func (h *Handler) userOverrideFields() g.Fields {
	return g.Fields{
		"id": &g.Field{Type: g.String},

		"add_user_id":    &g.Field{Type: g.String},
		"remove_user_id": &g.Field{Type: g.String},

		"add_user_name": &g.Field{
			Type: g.String,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				o, err := getUserOverride(p.Source)
				if err != nil {
					return nil, err
				}
				if o.AddUserID == "" {
					return nil, nil
				}
				scrub := newScrubber(p.Context).scrub

				u, err := h.c.UserStore.FindOne(p.Context, o.AddUserID)
				if err != nil {
					return scrub(nil, err)
				}
				return u.Name, nil
			},
		},
		"remove_user_name": &g.Field{
			Type: g.String,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				o, err := getUserOverride(p.Source)
				if err != nil {
					return nil, err
				}
				if o.RemoveUserID == "" {
					return nil, nil
				}
				scrub := newScrubber(p.Context).scrub

				u, err := h.c.UserStore.FindOne(p.Context, o.RemoveUserID)
				if err != nil {
					return scrub(nil, err)
				}
				return u.Name, nil
			},
		},
		"start_time": &g.Field{Type: ISOTimestamp},
		"end_time":   &g.Field{Type: ISOTimestamp},

		"target_id": &g.Field{
			Type: g.String,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				o, err := getUserOverride(p.Source)
				if err != nil {
					return nil, err
				}
				return o.Target.TargetID(), nil
			},
		},
		"target_type": &g.Field{
			Type: assignmentTargetType,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				o, err := getUserOverride(p.Source)
				if err != nil {
					return nil, err
				}
				return o.Target.TargetType(), nil
			},
		},
	}
}

func (h *Handler) updateUserOverrideField() *g.Field {
	return &g.Field{
		Type: h.userOverride,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "UpdateUserOverrideInput",
					Fields: g.InputObjectConfigFieldMap{
						"id":          &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"target_id":   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"target_type": &g.InputObjectFieldConfig{Type: g.NewNonNull(userOverrideTargetType)},

						"add_user_id":    &g.InputObjectFieldConfig{Type: g.String},
						"remove_user_id": &g.InputObjectFieldConfig{Type: g.String},

						"start_time": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"end_time":   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
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

			var o override.UserOverride

			var tgt assignment.RawTarget
			tgt.ID, _ = m["target_id"].(string)
			tgt.Type, _ = m["target_type"].(assignment.TargetType)
			o.Target = tgt
			o.ID, _ = m["id"].(string)
			o.AddUserID, _ = m["add_user_id"].(string)
			o.RemoveUserID, _ = m["remove_user_id"].(string)

			startStr, _ := m["start_time"].(string)
			endStr, _ := m["end_time"].(string)

			var err error
			o.Start, err = time.Parse(time.RFC3339, startStr)
			if err != nil {
				return nil, validation.NewFieldError("Start", err.Error())
			}
			o.End, err = time.Parse(time.RFC3339, endStr)
			if err != nil {
				return nil, validation.NewFieldError("End", err.Error())
			}

			return scrub(o, h.c.OverrideStore.UpdateUserOverride(p.Context, &o))
		},
	}
}
