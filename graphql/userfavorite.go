package graphql

import (
	"errors"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/permission"

	g "github.com/graphql-go/graphql"
)

func (h *Handler) setUserFavoriteField() *g.Field {
	return &g.Field{
		Type: h.assignmentTarget,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Description: "The target to set as a favorite.",
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "SetFavoriteInput",
					Fields: g.InputObjectConfigFieldMap{
						"target_type": &g.InputObjectFieldConfig{Type: g.NewNonNull(assignmentTargetType)},
						"target_id":   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
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
			tgt.Type, _ = m["target_type"].(assignment.TargetType)
			tgt.ID, _ = m["target_id"].(string)

			userID := permission.UserID(p.Context)

			err := h.c.UserFavoriteStore.Set(p.Context, userID, tgt)
			return newScrubber(p.Context).scrub(tgt, err)
		},
	}
}

func (h *Handler) unsetUserFavoriteField() *g.Field {
	return &g.Field{
		Type: h.assignmentTarget,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Description: "The target to unset as a favorite.",
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "UnsetFavoriteInput",
					Fields: g.InputObjectConfigFieldMap{
						"target_type": &g.InputObjectFieldConfig{Type: g.NewNonNull(assignmentTargetType)},
						"target_id":   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
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
			tgt.Type, _ = m["target_type"].(assignment.TargetType)
			tgt.ID, _ = m["target_id"].(string)

			userID := permission.UserID(p.Context)
			err := h.c.UserFavoriteStore.Unset(p.Context, userID, &tgt)
			return newScrubber(p.Context).scrub(tgt, err)
		},
	}
}
