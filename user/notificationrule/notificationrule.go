package notificationrule

import (
	"github.com/google/uuid"
	"github.com/target/goalert/validation/validate"
)

type NotificationRule struct {
	ID              string    `json:"id"`
	UserID          string    `json:"-"`
	DelayMinutes    int       `json:"delay"`
	ContactMethodID uuid.UUID `json:"contact_method_id"`
}

func validateDelay(d int) error {
	return validate.Range("DelayMinutes", d, 0, 9000)
}

func (n NotificationRule) Normalize(update bool) (*NotificationRule, error) {
	err := validateDelay(n.DelayMinutes)

	if !update {
		err = validate.Many(
			err,
			validate.UUID("UserID", n.UserID),
		)
	}
	if err != nil {
		return nil, err
	}

	return &n, nil
}
