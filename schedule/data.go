package schedule

import "time"

// Data contains configuration for a single schedule.
type Data struct {
	V1 struct{ TemporarySchedules []FixedShiftGroup }
}

// TempOnCall will calculate any on-call users for the given time. isActive will
// be true if a temporary schedule is active.
func (data *Data) TempOnCall(t time.Time) (isActive bool, users []string) {
	if data == nil {
		return false, nil
	}

	for _, grp := range data.V1.TemporarySchedules {
		if t.Before(grp.Start) || !t.Before(grp.End) {
			continue
		}
		isActive = true
		for _, shift := range grp.Shifts {
			if t.Before(shift.Start) || !t.Before(shift.End) {
				continue
			}
			users = append(users, shift.UserID)
		}

		// only one group will ever be active (should be merged & sorted)
		break
	}

	return isActive, users
}
