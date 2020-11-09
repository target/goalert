package timeutil

import "time"

// PrevWeekday will return t at the start of the most recent
// weekday w. The returned value will be identical if it is already
// the start of the requested weekday.
func PrevWeekday(t time.Time, w time.Weekday) time.Time {
	year, month, day := t.Date()
	adj := int(w - t.Weekday())
	if adj < 1 {
		adj += 7
	}
	adj -= 7
	if adj == 0 {
		return t
	}
	return time.Date(year, month, day+adj, 0, 0, 0, 0, t.Location())
}

// NextWeekday will return t at the start of the next
// weekday w. The returned value will always be in the future.
func NextWeekday(t time.Time, w time.Weekday) time.Time {
	year, month, day := t.Date()
	adj := int(w - t.Weekday())
	if adj < 1 {
		adj += 7
	}
	return time.Date(year, month, day+adj, 0, 0, 0, 0, t.Location())
}
