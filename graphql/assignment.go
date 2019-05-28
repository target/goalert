package graphql

import (
	"fmt"
	"github.com/target/goalert/assignment"
	"strconv"

	g "github.com/graphql-go/graphql"
)

var assignmentSourceType = g.NewEnum(g.EnumConfig{
	Name: "AssignmentSourceType",
	Values: g.EnumValueConfigMap{
		"alert":                  &g.EnumValueConfig{Value: assignment.SrcTypeAlert},
		"escalation_policy_step": &g.EnumValueConfig{Value: assignment.SrcTypeEscalationPolicyStep},
		"rotation_participant":   &g.EnumValueConfig{Value: assignment.SrcTypeRotationParticipant},
		"schedule_rule":          &g.EnumValueConfig{Value: assignment.SrcTypeScheduleRule},
		"service":                &g.EnumValueConfig{Value: assignment.SrcTypeService},
		"user":                   &g.EnumValueConfig{Value: assignment.SrcTypeUser},
	},
})

var assignmentTargetType = g.NewEnum(g.EnumConfig{
	Name: "AssignmentTargetType",
	Values: g.EnumValueConfigMap{
		"escalation_policy":   &g.EnumValueConfig{Value: assignment.TargetTypeEscalationPolicy},
		"notification_policy": &g.EnumValueConfig{Value: assignment.TargetTypeNotificationPolicy},
		"rotation":            &g.EnumValueConfig{Value: assignment.TargetTypeRotation},
		"service":             &g.EnumValueConfig{Value: assignment.TargetTypeService},
		"schedule":            &g.EnumValueConfig{Value: assignment.TargetTypeSchedule},
		"user":                &g.EnumValueConfig{Value: assignment.TargetTypeUser},
	},
})

func targetTypeField(t assignment.TargetType) *g.Field {
	return &g.Field{
		Type:    assignmentTargetType,
		Resolve: func(p g.ResolveParams) (interface{}, error) { return t, nil },
	}
}

func sourceTypeField(t assignment.SrcType) *g.Field {
	return &g.Field{
		Type:    assignmentSourceType,
		Resolve: func(p g.ResolveParams) (interface{}, error) { return t, nil },
	}
}

func (h *Handler) sourceField() *g.Field {
	return &g.Field{
		Type: h.sourceType,
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			if t, ok := p.Source.(assignment.Source); ok {
				switch t.SourceType() {
				case assignment.SrcTypeAlert:
					id, err := strconv.Atoi(t.SourceID())
					if err != nil {
						return nil, err
					}
					return newScrubber(p.Context).scrub(h.c.AlertStore.FindOne(p.Context, id))
				case assignment.SrcTypeEscalationPolicyStep:
					return newScrubber(p.Context).scrub(h.c.EscalationStore.FindOneStep(p.Context, t.SourceID()))
				case assignment.SrcTypeRotationParticipant:
					return newScrubber(p.Context).scrub(h.c.RotationStore.FindParticipant(p.Context, t.SourceID()))
				case assignment.SrcTypeScheduleRule:
				case assignment.SrcTypeService:
					return newScrubber(p.Context).scrub(h.c.ServiceStore.FindOne(p.Context, t.SourceID()))
				case assignment.SrcTypeUser:
					return newScrubber(p.Context).scrub(h.c.UserStore.FindOne(p.Context, t.SourceID()))
				}
			}
			return p.Source, nil
		},
	}
}

func (h *Handler) targetField() *g.Field {
	return &g.Field{
		Type: h.targetType,
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			if t, ok := p.Source.(assignment.Target); ok {
				switch t.TargetType() {
				case assignment.TargetTypeEscalationPolicy:
					return newScrubber(p.Context).scrub(h.c.EscalationStore.FindOnePolicy(p.Context, t.TargetID()))
				case assignment.TargetTypeNotificationPolicy:
				case assignment.TargetTypeRotation:
					return newScrubber(p.Context).scrub(h.c.RotationStore.FindRotation(p.Context, t.TargetID()))
				case assignment.TargetTypeSchedule:
					return newScrubber(p.Context).scrub(h.c.ScheduleStore.FindOne(p.Context, t.TargetID()))
				case assignment.TargetTypeService:
					return newScrubber(p.Context).scrub(h.c.ServiceStore.FindOne(p.Context, t.TargetID()))
				case assignment.TargetTypeUser:
					return newScrubber(p.Context).scrub(h.c.UserStore.FindOne(p.Context, t.TargetID()))
				}
			}
			return p.Source, nil
		},
	}
}

func (h *Handler) assignmentSourceFields() g.Fields {
	return g.Fields{
		"source_id": &g.Field{
			Type: g.String,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				if s, ok := p.Source.(assignment.Source); ok {
					return s.SourceID(), nil
				}
				return nil, fmt.Errorf("invalid source type %T", p.Source)
			},
		},
		"source_type": &g.Field{
			Type: assignmentSourceType,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				if s, ok := p.Source.(assignment.Source); ok {
					return s.SourceType(), nil
				}
				return nil, fmt.Errorf("invalid source type %T", p.Source)
			},
		},
		"source": h.sourceField(),
	}
}

func (h *Handler) assignmentTargetFields() g.Fields {
	return g.Fields{
		"target_id": &g.Field{
			Type: g.String,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				if s, ok := p.Source.(assignment.Target); ok {
					return s.TargetID(), nil
				}
				return nil, fmt.Errorf("invalid target type %T", p.Source)
			},
		},
		"target_type": &g.Field{
			Type: assignmentTargetType,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				if s, ok := p.Source.(assignment.Target); ok {
					return s.TargetType(), nil
				}
				return nil, fmt.Errorf("invalid target type %T", p.Source)
			},
		},
		"target": h.targetField(),
		"target_name": &g.Field{
			Type: g.String,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				var scrub = newScrubber(p.Context).scrub
				if t, ok := p.Source.(assignment.TargetNamer); ok && t.TargetName() != "" {
					return t.TargetName(), nil
				}

				if t, ok := p.Source.(assignment.Target); ok {
					switch t.TargetType() {
					case assignment.TargetTypeUser:
						tgt, err := h.c.UserStore.FindOne(p.Context, t.TargetID())
						if err != nil {
							return scrub(nil, err)
						}
						return tgt.Name, nil
					case assignment.TargetTypeRotation:
						tgt, err := h.c.RotationStore.FindRotation(p.Context, t.TargetID())
						if err != nil {
							return scrub(nil, err)
						}
						return tgt.Name, nil
					case assignment.TargetTypeSchedule:
						tgt, err := h.c.ScheduleStore.FindOne(p.Context, t.TargetID())
						if err != nil {
							return scrub(nil, err)
						}
						return tgt.Name, nil
					}
				}
				return nil, nil
			},
		},
	}
}
