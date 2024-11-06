package schedule

import (
	"time"

	"github.com/google/uuid"
)

// Data contains configuration for a single schedule.
type Data struct {
	V1 struct {
		TemporarySchedules      []TemporarySchedule
		OnCallNotificationRules []OnCallNotificationRule
	}
}

// TempOnCall will calculate any on-call users for the given time. isActive will
// be true if a temporary schedule is active.
func (data *Data) TempOnCall(t time.Time) (isActive bool, users []uuid.UUID) {
	if data == nil {
		return false, nil
	}

	for _, temp := range data.V1.TemporarySchedules {
		if t.Before(temp.Start) || !t.Before(temp.End) {
			continue
		}
		isActive = true
		for _, shift := range temp.Shifts {
			if t.Before(shift.Start) || !t.Before(shift.End) {
				continue
			}
			id, err := uuid.Parse(shift.UserID)
			if err != nil {
				continue
			}
			users = append(users, id)
		}

		// only one TemporarySchedule will ever be active (should be merged & sorted)
		break
	}

	return isActive, users
}
