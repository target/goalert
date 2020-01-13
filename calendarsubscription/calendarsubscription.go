package calendarsubscription

import (
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/validation/validate"
)

type CalendarSubscription struct {
	ID         string
	Name       string
	UserID     string
	ScheduleID string
	LastAccess time.Time
	Disabled   bool

	// Config provides necessary parameters CalendarSubscription Config (i.e. ReminderMinutes)
	Config struct {
		ReminderMinutes []int
	}
}

func (cs CalendarSubscription) Normalize() (*CalendarSubscription, error) {
	if cs.ID == "" {
		cs.ID = uuid.NewV4().String()
	}

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
