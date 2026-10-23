package timeutil

import (
	"sort"
	"time"
)

// StartOfDay will return the start of the day in t's location.
//
// If local midnight does not exist on that day (a DST transition at midnight),
// the first instant after the gap is returned.
func StartOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return startOfDay(y, m, d, t.Location())
}

// AddDays returns the time with the same wall-clock value as t, on the calendar
// day `days` days from t's calendar day, in t's location.
//
// Unlike t.AddDate(0, 0, days), the result is always on the expected calendar
// day: if the wall-clock value does not exist on that day (a DST transition),
// the first instant after it is returned, and if it exists twice, the first
// occurrence is returned.
func AddDays(t time.Time, days int) time.Time {
	y, m, d := t.Date()
	h, min, sec := t.Clock()
	c := Clock(time.Duration(h)*time.Hour + time.Duration(min)*time.Minute + time.Duration(sec)*time.Second + time.Duration(t.Nanosecond()))
	return c.FirstOfDay(startOfDay(y, m, d+days, t.Location()))
}

// startOfDay returns the earliest instant of the given calendar day in loc.
//
// The day value is normalized like time.Date (e.g., day 0 is the last day of
// the previous month).
//
// It is equivalent to time.Date(year, month, day, 0, 0, 0, 0, loc) for every
// day where local midnight exists. When it does not (a DST transition at
// midnight, e.g. America/Santiago or America/Havana) time.Date may return an
// instant on the previous day; startOfDay instead returns the first instant
// after the gap, which is the earliest instant of the requested day.
func startOfDay(year int, month time.Month, day int, loc *time.Location) time.Time {
	// Normalize the calendar date via noon, which exists on every day.
	y, m, d := time.Date(year, month, day, 12, 0, 0, 0, loc).Date()

	t := time.Date(y, m, d, 0, 0, 0, 0, loc)
	if ty, tm, td := t.Date(); ty == y && tm == m && td == d {
		return t
	}

	// Local midnight does not exist (a DST gap spans it) and time.Date
	// resolved it to an instant on the previous day, before the gap. The
	// earliest instant of the requested day is the instant the gap begins,
	// i.e. the first instant after t with a different zone offset.
	_, offset := t.Zone()
	mins := sort.Search(24*60, func(min int) bool {
		_, o := t.Add(time.Duration(min) * time.Minute).Zone()
		return o != offset
	})
	return t.Add(time.Duration(mins) * time.Minute)
}
