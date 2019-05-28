package graphql

import (
	"fmt"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/schedule/shiftcalc"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation/validate"

	g "github.com/graphql-go/graphql"
	"github.com/pkg/errors"
)

func getScheduleAssignment(src interface{}) (*shiftcalc.ScheduleAssignment, error) {
	switch s := src.(type) {
	case *shiftcalc.ScheduleAssignment:
		return s, nil
	case shiftcalc.ScheduleAssignment:
		return &s, nil
	default:
		return nil, fmt.Errorf("could not id of user (unknown source type %T)", s)
	}
}

func (h *Handler) scheduleAssignmentFields() g.Fields {
	return g.Fields{
		"id": &g.Field{
			Type: g.String,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				asn, err := getScheduleAssignment(p.Source)
				if err != nil {
					return nil, err
				}
				return fmt.Sprintf("Schedule(%s)/%s(%s)", asn.ScheduleID, asn.Target.TargetType(), asn.Target.TargetID()), nil
			},
		},
		"schedule_id": &g.Field{Type: g.String},
		"target_id": &g.Field{
			Type: g.String,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				asn, err := getScheduleAssignment(p.Source)
				if err != nil {
					return nil, err
				}
				return asn.Target.TargetID(), nil
			},
		},
		"target_type": &g.Field{
			Type: g.String,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				asn, err := getScheduleAssignment(p.Source)
				if err != nil {
					return nil, err
				}
				return asn.Target.TargetType().String(), nil
			},
		},
		"rotation": &g.Field{
			Type: h.rotation,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				asn, err := getScheduleAssignment(p.Source)
				if err != nil {
					return nil, err
				}

				if asn.Target.TargetType() != assignment.TargetTypeRotation {
					return nil, nil
				}

				return newScrubber(p.Context).scrub(h.c.RotationStore.FindRotation(p.Context, asn.Target.TargetID()))
			},
		},
		"user": &g.Field{
			Type: h.user,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				asn, err := getScheduleAssignment(p.Source)
				if err != nil {
					return nil, err
				}

				if asn.Target.TargetType() != assignment.TargetTypeUser {
					return nil, nil
				}

				return newScrubber(p.Context).scrub(h.c.UserStore.FindOne(p.Context, asn.Target.TargetID()))
			},
		},
		"rules":  &g.Field{Type: g.NewList(h.scheduleRule)},
		"shifts": &g.Field{Type: g.NewList(h.scheduleShift)},
	}
}

func (h *Handler) deleteScheduleAssignmentField() *g.Field {
	return &g.Field{
		Type: h.schedule,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "DeleteScheduleAssignmentInput",
					Fields: g.InputObjectConfigFieldMap{
						"schedule_id": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"target_id":   &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
						"target_type": &g.InputObjectFieldConfig{Type: g.NewNonNull(assignmentTargetType)},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			schedID, _ := m["schedule_id"].(string)
			err := validate.UUID("ScheduleID", schedID)
			if err != nil {
				return nil, err
			}

			var ctx = log.WithField(p.Context, "ScheduleID", schedID)

			var asnTarget assignment.RawTarget
			asnTarget.ID, _ = m["target_id"].(string)
			asnTarget.Type, _ = m["target_type"].(assignment.TargetType)

			err = validate.UUID("TargetID", asnTarget.ID)
			if err != nil {
				return nil, err
			}

			err = h.c.ScheduleRuleStore.DeleteByTarget(ctx, schedID, asnTarget)
			if err != nil {
				return newScrubber(ctx).scrub(nil, err)
			}

			return newScrubber(ctx).scrub(h.c.ScheduleStore.FindOne(ctx, schedID))
		},
	}
}
