package assignment

import "strconv"

// Source contains information about the source, or subject of an assignment.
type Source interface {
	SourceType() SrcType
	SourceID() string
}
type RawSource struct {
	Type SrcType
	ID   string
}

func NewRawSource(s Source) RawSource {
	return RawSource{Type: s.SourceType(), ID: s.SourceID()}
}
func (rt RawSource) SourceType() SrcType {
	return rt.Type
}
func (rt RawSource) SourceID() string {
	return rt.ID
}

type (
	// AlertSource implements the Source interface by wrapping an Alert ID.
	AlertSource int
	// EscalationPolicyStepSource implements the Source interface by wrapping an EsclationPolicyStep ID.
	EscalationPolicyStepSource string
	// RotationParticipantSource implements the Source interface by wrapping a RotationParticipant ID.
	RotationParticipantSource string
	// ScheduleRuleSource implements the Source interface by wrapping a ScheduleRule ID.
	ScheduleRuleSource string
	// ServiceSource implements the Source interface by wrapping a Service ID.
	ServiceSource string
	// UserSource implements the Source interface by wrapping a UserSource ID.
	UserSource string
)

// SourceType implements the Source interface.
func (AlertSource) SourceType() SrcType { return SrcTypeAlert }

// SourceID implements the Source interface.
func (a AlertSource) SourceID() string { return strconv.Itoa(int(a)) }

// SourceType implements the Source interface.
func (EscalationPolicyStepSource) SourceType() SrcType { return SrcTypeEscalationPolicyStep }

// SourceID implements the Source interface.
func (e EscalationPolicyStepSource) SourceID() string { return string(e) }

// SourceType implements the Source interface.
func (RotationParticipantSource) SourceType() SrcType { return SrcTypeRotationParticipant }

// SourceID implements the Source interface.
func (r RotationParticipantSource) SourceID() string { return string(r) }

// SourceType implements the Source interface.
func (ScheduleRuleSource) SourceType() SrcType { return SrcTypeScheduleRule }

// SourceID implements the Source interface.
func (s ScheduleRuleSource) SourceID() string { return string(s) }

// SourceType implements the Source interface.
func (ServiceSource) SourceType() SrcType { return SrcTypeService }

// SourceID implements the Source interface.
func (s ServiceSource) SourceID() string { return string(s) }

// SourceType implements the Source interface.
func (UserSource) SourceType() SrcType { return SrcTypeUser }

// SourceID implements the Source interface.
func (u UserSource) SourceID() string { return string(u) }
