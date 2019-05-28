package rotation

import (
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/validation/validate"
)

type Participant struct {
	ID         string `json:"id"`
	Position   int    `json:"position"`
	RotationID string `json:"rotation_id"`
	Target     assignment.Target
}

func (p Participant) Normalize() (*Participant, error) {
	err := validate.Many(
		validate.UUID("RotationID", p.RotationID),
		validate.UUID("TargetID", p.Target.TargetID()),
		validate.OneOf("TargetType", p.Target.TargetType(),
			assignment.TargetTypeUser,
		),
		validate.Range("Position", p.Position, 0, 9000),
	)

	if err != nil {
		return nil, err
	}
	return &p, nil
}
