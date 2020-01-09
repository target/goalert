package calendarsubscription

import (
	"github.com/target/goalert/validation/validate"
	"time"
)

type CalendarSubscription struct {
	ID         string
	Name       string
	UserID     string
	LastAccess time.Time
	Disabled   bool
	Config     []byte
	ScheduleID string

	ReminderMinutes []int
}

func (cs CalendarSubscription) Normalize() (*CalendarSubscription, error) {
	err := validate.Many(
		validate.IDName("Name", cs.Name),
		validate.UUID("ID", cs.ID),
		validate.UUID("UserID", cs.UserID),
	)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}
