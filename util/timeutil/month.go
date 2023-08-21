package timeutil

import "time"

// MonthBeginning will return the start of the month where time t falls.
func MonthBeginning(t time.Time, l time.Location) time.Time {
	year, month, _ := t.Date()
	firstday := time.Date(year, month, 1, 0, 0, 0, 0, &l)
	return firstday
}

// MonthEnd will return the end of the month where time t falls.
func MonthEnd(t time.Time, l time.Location) time.Time {
	lastday := MonthBeginning(t, l).AddDate(0, 1, 0).Add(time.Nanosecond * -1)
	return lastday
}
