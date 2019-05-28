package override

import (
	"fmt"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
	"time"
)

// A UserOverride is used to add, remove, or change which user is on call.
type UserOverride struct {
	ID           string    `json:"id,omitempty"`
	AddUserID    string    `json:"add_user_id,omitempty"`
	RemoveUserID string    `json:"remove_user_id,omitempty"`
	Start        time.Time `json:"start_time,omitempty"`
	End          time.Time `json:"end_time,omitempty"`
	Target       assignment.Target
}

const debugTimeFmt = "MonJan2_2006@3:04pm"

func (o UserOverride) String() string {
	var tgt string
	if o.Target != nil {
		tgt = ", " + o.Target.TargetType().String() + "(" + o.Target.TargetID() + ")"
	}
	return fmt.Sprintf("UserOverride{Start: %s, End: %s, AddUserID: %s, RemoveUserID: %s%s}",
		o.Start.Local().Format(debugTimeFmt),
		o.End.Local().Format(debugTimeFmt),
		o.AddUserID,
		o.RemoveUserID,
		tgt,
	)
}

// Normalize will validate fields and return a normalized copy.
func (o UserOverride) Normalize() (*UserOverride, error) {
	var err error
	if o.AddUserID == "" && o.RemoveUserID == "" {
		err = validation.NewFieldError("UserID", "must specify AddUserID and/or RemoveUserID")
	}
	if o.AddUserID != "" {
		err = validate.Many(err, validate.UUID("AddUserID", o.AddUserID))
	}
	if o.RemoveUserID != "" {
		err = validate.Many(err, validate.UUID("RemoveUserID", o.RemoveUserID))
	}
	if !o.Start.Before(o.End) {
		err = validate.Many(err, validation.NewFieldError("End", "must occur after Start time"))
	}
	err = validate.Many(err,
		validate.UUID("TargetID", o.Target.TargetID()),
		validate.OneOf("TargetType", o.Target.TargetType(), assignment.TargetTypeSchedule),
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}
