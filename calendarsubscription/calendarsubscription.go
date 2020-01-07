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
	Config	   []byte  `json:"config"`
	ScheduleID string 	`json:"schedule_id"`

	NMinutes []int `json:"notification_minutes"`

}

func (cs CalendarSubscription) Normalize() (*CalendarSubscription, error) {
	err := validate.Many(
		validate.IDName("CalendarSubscriptionName", cs.Name),
		validate.UUID("CalendarSubscriptionID", cs.ID),
		validate.UUID("CalendarSubscriptionUserID", cs.UserID),
	)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}
