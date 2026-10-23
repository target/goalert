package timeutil

import (
	"time"
)

// AddClock will add the provided number of clock hours to the given time, taking
// things like DST into account.
func AddClock(t time.Time, c Clock) time.Time {
	c += NewClockFromTime(t)
	days, c := c.Days()

	y, m, d := t.Date()
	return c.FirstOfDay(startOfDay(y, m, d+days, t.Location()))
}

// HoursBetween will return the number of full hours from a to b with
// respect to clock time and DST.
//
// It is assumed a and b are of the same location.
func HoursBetween(a, b time.Time) int {
	diff := b.Sub(a)

	_, offsetA := a.Zone()
	_, offsetB := b.Zone()
	if offsetA == offsetB {
		return int(diff / time.Hour)
	}
	diff += time.Duration(offsetB-offsetA) * time.Second
	return int(diff / time.Hour)
}

// ClockDiff will return the amount of clock time from a to b with
// respect to DST.
//
// It is assumed a and b are of the same location.
func ClockDiff(a, b time.Time) Clock {
	diff := b.Sub(a)

	_, offsetA := a.Zone()
	_, offsetB := b.Zone()
	if offsetA == offsetB {
		return Clock(diff)
	}
	diff += time.Duration(offsetB-offsetA) * time.Second
	return Clock(diff)
}
