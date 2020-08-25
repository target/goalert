package schedule

import "time"

type ScheduleData struct {
	V1 struct{ TemporarySchedules []FixedShiftGroup }
}

func (data *ScheduleData) TempOnCall(t time.Time) (bool, []string) {
	if data == nil {
		return false, nil
	}
	var users []string
	for _, grp := range data.V1.TemporarySchedules {
		if t.Before(grp.Start) || !t.Before(grp.End) {
			continue
		}

		for _, shift := range grp.Shifts {
			if t.Before(shift.Start) || !t.Before(shift.End) {
				continue
			}
			users = append(users, shift.UserID)
		}

		// only one group will ever be active (should be merged & sorted)
		break
	}

	return true, users
}
