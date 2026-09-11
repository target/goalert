package label

import (
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/validation/validate"
)

// A Label is a key-value pair assigned to a target.
type Label struct {
	Target assignment.Target
	Key    string `json:"key"`
	Value  string `json:"value"`
}

// Normalize will validate and normalize the label, returning a copy.
func (l Label) Normalize() (*Label, error) {
	return &l, validate.Many(
		validate.OneOf("TargetType", l.Target.TargetType(),
			assignment.TargetTypeService,
			assignment.TargetTypeEscalationPolicy,
			assignment.TargetTypeSchedule,
			assignment.TargetTypeRotation,
		),
		validate.UUID("TargetID", l.Target.TargetID()),
		validate.LabelKey("Key", l.Key),
		validate.LabelValue("Value", l.Value),
	)
}
