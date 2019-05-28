package graphql

import (
	"fmt"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation"
	"time"

	g "github.com/graphql-go/graphql"
	"github.com/pkg/errors"
)

var rotationTypeEnum = g.NewEnum(g.EnumConfig{
	Name: "RotationType",
	Values: g.EnumValueConfigMap{
		"daily":  &g.EnumValueConfig{Value: rotation.TypeDaily},
		"weekly": &g.EnumValueConfig{Value: rotation.TypeWeekly},
		"hourly": &g.EnumValueConfig{Value: rotation.TypeHourly},
	},
})

func (h *Handler) rotationShiftFields() g.Fields {
	return g.Fields{
		"start_time":     &g.Field{Type: ISOTimestamp},
		"end_time":       &g.Field{Type: ISOTimestamp},
		"participant_id": &g.Field{Type: g.String},
	}
}

func getRot(src interface{}) (*rotation.Rotation, error) {
	switch s := src.(type) {
	case rotation.Rotation:
		return &s, nil
	case *rotation.Rotation:
		return s, nil
	default:
		return nil, fmt.Errorf("invalid source type %T for rotation", s)
	}
}

func (h *Handler) rotationFields() g.Fields {
	return g.Fields{
		"id":           &g.Field{Type: g.String},
		"name":         &g.Field{Type: g.String},
		"description":  &g.Field{Type: g.String},
		"type":         &g.Field{Type: rotationTypeEnum},
		"start":        &g.Field{Type: ISOTimestamp},
		"shift_length": &g.Field{Type: g.Int},
		"target_type":  targetTypeField(assignment.TargetTypeRotation),
		"schedule_id": &g.Field{
			Type: g.String,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				r, err := getRot(p.Source)
				if err != nil {
					return nil, err
				}

				scrub := newScrubber(p.Context).scrub
				return scrub(h.legacyDB.ScheduleIDFromRotation(p.Context, r.ID))
			},
		},
		"schedule": &g.Field{
			Type: g.String,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				r, err := getRot(p.Source)
				if err != nil {
					return nil, err
				}

				scrub := newScrubber(p.Context).scrub

				schedID, err := h.legacyDB.ScheduleIDFromRotation(p.Context, r.ID)
				if err != nil {
					return scrub(nil, err)
				}

				return scrub(h.c.ScheduleStore.FindOne(p.Context, schedID))
			},
		},

		"shifts": &g.Field{
			Type: g.NewList(h.rotationShift),
			Args: g.FieldConfigArgument{
				"start_time": &g.ArgumentConfig{Type: g.NewNonNull(g.String)},
				"end_time":   &g.ArgumentConfig{Type: g.NewNonNull(g.String)},
			},
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				return nil, nil
			},
		},

		"time_zone": &g.Field{
			Type: g.String,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				r, err := getRot(p.Source)
				if err != nil {
					return nil, err
				}

				return r.Start.Location().String(), nil
			},
		},

		"active_participant_id": &g.Field{
			Type: g.String,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				r, err := getRot(p.Source)
				if err != nil {
					return nil, err
				}
				scrub := newScrubber(p.Context).scrub

				state, err := h.c.RotationStore.State(p.Context, r.ID)
				if err == rotation.ErrNoState {
					return -1, nil
				}
				if err != nil {
					return scrub(nil, err)
				}

				return state.ParticipantID, nil
			},
		},

		"next_handoff_times": &g.Field{
			Type: g.NewList(ISOTimestamp),
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				r, err := getRot(p.Source)
				if err != nil {
					return nil, err
				}

				scrub := newScrubber(p.Context).scrub
				parts, err := h.c.RotationStore.FindAllParticipants(p.Context, r.ID)
				if err != nil {
					return scrub(nil, err)
				}

				shiftState, err := h.c.RotationStore.State(p.Context, r.ID)
				if err == rotation.ErrNoState {
					// rotation hasn't been started/processed yet
					return nil, nil
				}

				if err != nil {
					return scrub(nil, err)
				}

				var shifts []time.Time
				cEnd := r.EndTime(shiftState.ShiftStart)
				for range parts {
					shifts = append(shifts, cEnd)
					cEnd = r.EndTime(cEnd)
				}

				return shifts, nil
			},
		},

		"participants": &g.Field{
			Type: g.NewList(h.rotationParticipant),
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				r, err := getRot(p.Source)
				if err != nil {
					return nil, err
				}

				return newScrubber(p.Context).scrub(h.c.RotationStore.FindAllParticipants(p.Context, r.ID))
			},
		},
	}
}

func (h *Handler) createOrUpdateRotationField() *g.Field {
	return &g.Field{
		Type: g.NewObject(g.ObjectConfig{
			Name: "CreateOrUpdateRotationOutput",
			Fields: g.Fields{
				"created":  &g.Field{Type: g.Boolean, Description: "Signifies if a new record was created."},
				"rotation": &g.Field{Type: h.rotation, Description: "The created or updated record."},
			},
		}),
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewNonNull(g.NewInputObject(g.InputObjectConfig{
					Name:        "CreateOrUpdateRotationInput",
					Description: "Add rotation to a schedule",
					Fields: g.InputObjectConfigFieldMap{
						"id":           &g.InputObjectFieldConfig{Type: g.String, Description: "Specifies an existing rotation to update."},
						"name":         &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"description":  &g.InputObjectFieldConfig{Type: g.String},
						"type":         &g.InputObjectFieldConfig{Type: g.NewNonNull(rotationTypeEnum)},
						"start":        &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"shift_length": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Int)},
						"schedule_id":  &g.InputObjectFieldConfig{Type: g.String},
						"time_zone":    &g.InputObjectFieldConfig{Type: g.String},
					},
				})),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			var r rotation.Rotation

			r.ID, _ = m["id"].(string)
			r.Name, _ = m["name"].(string)
			r.Description, _ = m["description"].(string)
			r.Type, _ = m["type"].(rotation.Type)

			sTime, _ := m["start"].(string)
			var err error
			r.Start, err = time.Parse(time.RFC3339, sTime)
			if err != nil {
				return nil, validation.NewFieldError("start", "invalid format for time value: "+err.Error())
			}

			r.ShiftLength, _ = m["shift_length"].(int)
			schedID, ok := m["schedule_id"].(string)
			if ok {
				sched, err := h.c.ScheduleStore.FindOne(p.Context, schedID)
				if err != nil {
					return newScrubber(p.Context).scrub(nil, errors.Wrap(err, "lookup schedule"))
				}
				r.Name = sched.Name + " Rotation"
				r.Start = r.Start.In(sched.TimeZone)
			} else {
				tz, _ := m["time_zone"].(string)
				if tz == "" {
					return nil, validation.NewFieldError("time_zone", "must not be empty")
				}
				loc, err := util.LoadLocation(tz)
				if err != nil {
					return nil, validation.NewFieldError("time_zone", "invalid time_zone: "+err.Error())
				}
				r.Start = r.Start.In(loc)
			}

			var create bool

			if r.ID == "" {
				create = true
				rot, err := h.c.RotationStore.CreateRotation(p.Context, &r)
				if err != nil {
					return newScrubber(p.Context).scrub(nil, errors.Wrap(err, "create rotation"))
				}
				r = *rot
			} else {
				err = h.c.RotationStore.UpdateRotation(p.Context, &r)
				if err != nil {
					return newScrubber(p.Context).scrub(nil, errors.Wrap(err, "update rotation"))
				}
			}

			var resp struct {
				Created  bool              `json:"created"`
				Rotation rotation.Rotation `json:"rotation"`
			}
			resp.Created = create
			resp.Rotation = r
			return resp, nil
		},
	}
}

func (h *Handler) deleteRotationField() *g.Field {
	return &g.Field{
		Type: g.NewObject(g.ObjectConfig{
			Name:   "DeleteRotationOutput",
			Fields: g.Fields{"deleted_id": &g.Field{Type: g.String}},
		}),
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewNonNull(g.NewInputObject(g.InputObjectConfig{
					Name: "DeleteRotationInput",
					Fields: g.InputObjectConfigFieldMap{
						"id": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
					},
				})),
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
			return newScrubber(p.Context).scrub(r, h.c.RotationStore.DeleteRotation(p.Context, r.ID))
		},
	}
}

func (h *Handler) rotationsField() *g.Field {
	return &g.Field{
		Type: g.NewList(h.rotation),
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			return newScrubber(p.Context).scrub(h.c.RotationStore.FindAllRotations(p.Context))
		},
	}
}
func (h *Handler) rotationField() *g.Field {
	return &g.Field{
		Type: h.rotation,
		Args: g.FieldConfigArgument{
			"id": &g.ArgumentConfig{
				Type: g.NewNonNull(g.String),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			id, _ := p.Args["id"].(string)
			return newScrubber(p.Context).scrub(h.c.RotationStore.FindRotation(p.Context, id))
		},
	}
}
