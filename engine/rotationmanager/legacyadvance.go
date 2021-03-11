package rotationmanager

import (
	"time"

	"github.com/target/goalert/schedule/rotation"
)

func addHours(t time.Time, n int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+n, t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func addHoursAlwaysInc(t time.Time, n int) time.Time {
	res := addHours(t, n)
	if n < 0 {
		for !res.Before(t) {
			n--
			res = addHours(t, n)
		}
	} else {
		for !res.After(t) {
			n++
			res = addHours(t, n)
		}
	}

	return res
}

func calcOldEndTime(rot *rotation.Rotation, shiftStart time.Time) time.Time {
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
			cTime = addHoursAlwaysInc(cTime, -rot.ShiftLength)
		}
		return last
	}

	for !cTime.After(t) {
		cTime = addHoursAlwaysInc(cTime, rot.ShiftLength)
	}

	return cTime
}
