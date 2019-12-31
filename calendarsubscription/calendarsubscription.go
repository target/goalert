package calendarsubscription

import (
	"github.com/target/goalert/validation/validate"
	"time"
)

type CalendarSubscription struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	UserID     string    `json:"user_id"`
	LastAccess time.Time `json:"last_access"`
	Disabled   bool      `json:"disabled"`
}

func (cs CalendarSubscription) Normalize() (*CalendarSubscription, error) {
	err := validate.Many(
		validate.IDName("Name", cs.Name),
	)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}
