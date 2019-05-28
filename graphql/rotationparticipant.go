package graphql

import (
	"errors"
	"fmt"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/validation/validate"

	g "github.com/graphql-go/graphql"
)

func getRotationPart(s interface{}) (*rotation.Participant, error) {
	switch p := s.(type) {
	case rotation.Participant:
		return &p, nil
	case *rotation.Participant:
		return p, nil
	default:
		return nil, fmt.Errorf("invalid source type %T for rotation participant", p)
	}
}

func (h *Handler) rotationParticipantFields() g.Fields {
	return g.Fields{
		"id":          &g.Field{Type: g.String},
		"position":    &g.Field{Type: g.Int},
		"user_id":     &g.Field{Type: g.String},
		"rotation_id": &g.Field{Type: g.String},

		"rotation": &g.Field{
			Type: h.rotation,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				rp, err := getRotationPart(p.Source)
				if err != nil {
					return nil, err
				}

				return newScrubber(p.Context).scrub(h.c.RotationStore.FindRotation(p.Context, rp.RotationID))
			},
		},

		"user": &g.Field{
			Type: h.user,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				rp, err := getRotationPart(p.Source)
				if err != nil {
					return nil, err
				}

				if rp.Target.TargetType() == assignment.TargetTypeUser {
					return newScrubber(p.Context).scrub(h.c.UserStore.FindOne(p.Context, rp.Target.TargetID()))
				}
				return nil, errors.New("no user assigned to that rotation slot")
			},
		},
	}
}

func (h *Handler) addRotationParticipantField() *g.Field {
	return &g.Field{
		Description:       "Adds a new participant to the end of a rotation. The same user can be added multiple times.",
		Type:              h.rotationParticipant,
		DeprecationReason: "use addRotationParticipant2 instead",
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "AddRotationParticipantInput",
					Fields: g.InputObjectConfigFieldMap{
						"rotation_id": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"user_id":     &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			rp := &rotation.Participant{}
			var err error
			rp.RotationID, _ = m["rotation_id"].(string)

			if id, ok := m["user_id"].(string); ok {
				err = validate.UUID("user_id", id)
				if err != nil {
					return nil, err
				}
				rp.Target = assignment.UserTarget(id)
			}

			return newScrubber(p.Context).scrub(h.c.RotationStore.AddParticipant(p.Context, rp))
		},
	}
}

func (h *Handler) deleteRotationParticipantField() *g.Field {
	return &g.Field{
		Description:       "Remove a participant from a rotation.",
		DeprecationReason: "use deleteRotationParticipant2 instead",
		Type: g.NewObject(g.ObjectConfig{
			Name: "DeleteRotationParticipantOutput",
			Fields: g.Fields{
				"deleted_id":  &g.Field{Type: g.String},
				"rotation_id": &g.Field{Type: g.String},
			},
		}),
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "DeleteRotationParticipantInput",
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
				ID  string `json:"deleted_id"`
				RID string `json:"rotation_id"`
			}
			r.ID, _ = m["id"].(string)
			var err error
			r.RID, err = h.c.RotationStore.RemoveParticipant(p.Context, r.ID)
			return newScrubber(p.Context).scrub(r, err)

		},
	}
}

func (h *Handler) moveRotationParticipantField() *g.Field {
	return &g.Field{
		Description:       "Moves a participant to new_position, automatically shifting other participants around.",
		DeprecationReason: "use moveRotationParticipant2 instead",
		Type:              g.NewList(h.rotationParticipant),
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "MoveRotationParticipantInput",
					Fields: g.InputObjectConfigFieldMap{
						"id":           &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"new_position": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Int)},
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
			newPos, _ := m["new_position"].(int)

			err := h.c.RotationStore.MoveParticipant(p.Context, id, newPos)
			if err != nil {
				return newScrubber(p.Context).scrub(nil, err)
			}

			rp, err := h.c.RotationStore.FindParticipant(p.Context, id)
			if err != nil {
				return newScrubber(p.Context).scrub(nil, err)
			}
			return newScrubber(p.Context).scrub(h.c.RotationStore.FindAllParticipants(p.Context, rp.RotationID))
		},
	}
}

func (h *Handler) addRotationParticipant2Field() *g.Field {
	return &g.Field{
		Description: "Adds a new participant to the end of a rotation. The same user can be added multiple times.",
		Type:        h.rotation,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "AddRotationParticipant2Input",
					Fields: g.InputObjectConfigFieldMap{
						"rotation_id": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"user_id":     &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			rp := &rotation.Participant{}
			rp.RotationID, _ = m["rotation_id"].(string)
			scrub := newScrubber(p.Context).scrub
			if id, ok := m["user_id"].(string); ok {
				err := validate.UUID("user_id", id)
				if err != nil {
					return nil, err
				}
				rp.Target = assignment.UserTarget(id)
			}

			rp, err := h.c.RotationStore.AddParticipant(p.Context, rp)
			if err != nil {
				return scrub(nil, err)
			}

			return scrub(h.c.RotationStore.FindRotation(p.Context, rp.RotationID))
		},
	}
}

func (h *Handler) deleteRotationParticipant2Field() *g.Field {
	return &g.Field{
		Description: "Remove a participant from a rotation.",
		Type:        h.rotation,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "DeleteRotationParticipant2Input",
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
			scrub := newScrubber(p.Context).scrub
			partID, _ := m["id"].(string)
			rotID, err := h.c.RotationStore.RemoveParticipant(p.Context, partID)
			if err != nil {
				return scrub(nil, err)
			}

			return scrub(h.c.RotationStore.FindRotation(p.Context, rotID))
		},
	}
}

func (h *Handler) moveRotationParticipant2Field() *g.Field {
	return &g.Field{
		Description: "Moves a participant to new_position, automatically shifting other participants around.",
		Type:        h.rotation,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "MoveRotationParticipant2Input",
					Fields: g.InputObjectConfigFieldMap{
						"id":           &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"new_position": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Int)},
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
			newPos, _ := m["new_position"].(int)
			scrub := newScrubber(p.Context).scrub
			err := h.c.RotationStore.MoveParticipant(p.Context, id, newPos)
			if err != nil {
				return scrub(nil, err)
			}

			rp, err := h.c.RotationStore.FindParticipant(p.Context, id)
			if err != nil {
				return scrub(nil, err)
			}

			return scrub(h.c.RotationStore.FindRotation(p.Context, rp.RotationID))
		},
	}
}

func (h *Handler) setActiveParticipantField() *g.Field {
	return &g.Field{
		Description: "Sets a specified participant as active in the provided rotation.",
		Type:        h.rotation,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "SetActiveParticipantInput",
					Fields: g.InputObjectConfigFieldMap{
						"rotation_id":    &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"participant_id": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			rotID, _ := m["rotation_id"].(string)
			partID, _ := m["participant_id"].(string)
			scrub := newScrubber(p.Context).scrub
			err := h.c.RotationStore.SetActiveParticipant(p.Context, rotID, partID)
			if err != nil {
				return scrub(nil, err)
			}

			return scrub(h.c.RotationStore.FindRotation(p.Context, rotID))
		},
	}
}
