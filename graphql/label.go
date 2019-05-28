package graphql

import (
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/label"

	g "github.com/graphql-go/graphql"
	"github.com/pkg/errors"
)

func (h *Handler) labelFields() g.Fields {
	return g.Fields{
		"key": &g.Field{
			Type: g.String,
		},
		"value": &g.Field{
			Type: g.String,
		},
	}
}

func (h *Handler) setLabelField() *g.Field {
	return &g.Field{
		Type: g.Boolean,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "SetLabelInput",
					Fields: g.InputObjectConfigFieldMap{
						"target_type": &g.InputObjectFieldConfig{Type: g.NewNonNull(assignmentTargetType)},
						"target_id":   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"key":         &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"value":       &g.InputObjectFieldConfig{Type: g.String},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			var lbl label.Label

			var tgt assignment.RawTarget
			tgt.ID, _ = m["target_id"].(string)
			tgt.Type, _ = m["target_type"].(assignment.TargetType)
			lbl.Target = tgt
			lbl.Key, _ = m["key"].(string)
			lbl.Value, _ = m["value"].(string)

			err := h.c.LabelStore.SetTx(p.Context, nil, &lbl)
			if err != nil {
				return newScrubber(p.Context).scrub(nil, errors.Wrap(err, "set label"))
			}

			return true, nil
		},
	}
}

func (h *Handler) labelKeysField() *g.Field {
	return &g.Field{
		Description: "All unique keys for labels.",
		Type:        g.NewList(g.String),
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			keys, err := h.c.LabelStore.UniqueKeys(p.Context)
			if err != nil {
				return newScrubber(p.Context).scrub(nil, errors.Wrap(err, "get unique keys"))
			}
			return newScrubber(p.Context).scrub(keys, nil)
		},
	}
}
