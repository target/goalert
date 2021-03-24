package rotationmanager

import (
	"time"

	"github.com/target/goalert/schedule/rotation"
)

// legacyAddHours is preserved from the old rotation calculation code to migrate from version 1 to 2.
// It simply reinterprets the date with the number of hours added to the clock time.
func legacyAddHours(t time.Time, n int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+n, t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// legacyAddHoursAlwaysInc is preserved from the old rotation calculation code to migrate
// from version 1 to 2. It adds a number of hours ensuring the unix timestamp always progresses
// forward.
func legacyAddHoursAlwaysInc(t time.Time, n int) time.Time {
	res := legacyAddHours(t, n)
	if n < 0 {
		for !res.Before(t) {
			n--
			res = legacyAddHours(t, n)
		}
	} else {
		for !res.After(t) {
			n++
			res = legacyAddHours(t, n)
		}
	}

	return res
}

func calcVersion1EndTime(rot *rotation.Rotation, shiftStart time.Time) time.Time {
	if rot.Type != rotation.TypeHourly {
		return rot.EndTime(shiftStart)
	}

	cTime := rot.Start.Truncate(time.Minute)
	t := shiftStart.In(cTime.Location())

	if cTime.After(t) {
		// reverse search
		last := cTime
		for cTime.After(t) {
			last = cTime
			cTime = legacyAddHoursAlwaysInc(cTime, -rot.ShiftLength)
		}
		return last
	}

	for !cTime.After(t) {
		cTime = legacyAddHoursAlwaysInc(cTime, rot.ShiftLength)
	}

	return cTime
}
