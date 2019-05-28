package assignment

// SrcType represents the source-type of an assignment.
type SrcType int

// Available SrcTypes
const (
	SrcTypeUnspecified SrcType = iota
	SrcTypeAlert
	SrcTypeEscalationPolicyStep
	SrcTypeRotationParticipant
	SrcTypeScheduleRule
	SrcTypeService
	SrcTypeUser
)

func (s SrcType) ParentType() TargetType {
	switch s {
	case SrcTypeEscalationPolicyStep:
		return TargetTypeEscalationPolicy
	case SrcTypeRotationParticipant:
		return TargetTypeRotation
	case SrcTypeScheduleRule:
		return TargetTypeSchedule
	case SrcTypeService:
		return TargetTypeService
	case SrcTypeUser:
		return TargetTypeUser
	}

	return TargetTypeUnspecified
}
