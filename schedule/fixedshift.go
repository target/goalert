package schedule

import (
	"sort"
	"time"
)

// A FixedShift represents an on-call user with a start and end time.
type FixedShift struct {
	Start, End time.Time
	UserID     string
}

// FixedOnCallUsers will return the users that would be on-call for the given configuration.
func FixedOnCallUsers(groups []FixedShiftGroup, t time.Time) (id []string, ok bool) {
	for _, grp := range groups {
		if !timeWithin(grp.Start, grp.End, t) {
			continue
		}

		// authoritive
		var onCall []string
		for _, s := range grp.Shifts {
			if !timeWithin(s.Start, s.End, t) {
				continue
			}
			onCall = append(onCall, s.UserID)
		}

		return dedup(onCall), true
	}

	return nil, false
}

func timeWithin(start, end, t time.Time) bool {
	if start.Before(t) {
		return false
	}

	return end.After(t)
}

func dedup(ids []string) []string {
	sort.Strings(ids)
	uniq := ids[:0]
	var last string
	for _, id := range ids {
		if id == last {
			continue
		}
		last = id
		uniq = append(uniq, id)
	}

	return uniq
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func clampShiftTimes(start, end time.Time, shifts []FixedShift) []FixedShift {
	result := shifts[:0]
	// trim/clamp shift times
	for _, s := range shifts {
		if s.Start.Before(start) {
			s.Start = start
		}
		if s.End.After(end) {
			s.End = end
		}
		if !s.End.After(s.Start) {
			continue
		}

		result = append(result, s)
	}
	return result
}

func mergeShiftsByTime(shifts []FixedShift) []FixedShift {
	if len(shifts) == 0 {
		return shifts
	}

	sort.Slice(shifts, func(i, j int) bool { return shifts[i].Start.Before(shifts[j].Start) })
	result := shifts[:1]
	for _, s := range shifts[1:] {
		l := len(result) - 1

		if !s.End.After(s.Start) {
			// omit empty time range
			continue
		}
		if s.Start.After(result[l].End) {
			result = append(result, s)
			continue
		}
		result[l].End = maxTime(result[l].End, s.End)
	}

	return result
}
func mergeShifts(shifts []FixedShift) []FixedShift {
	m := make(map[string][]FixedShift)
	for _, s := range shifts {
		m[s.UserID] = append(m[s.UserID], s)
	}
	result := shifts[:0]
	for _, s := range m {
		result = append(result, mergeShiftsByTime(s)...)
	}

	return result
}
