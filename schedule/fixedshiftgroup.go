package schedule

import (
	"sort"
	"time"
)

// FixedShiftGroup represents a timespan containing static pre-defined shifts of on-call users.
type FixedShiftGroup struct {
	Start, End time.Time
	Shifts     []FixedShift
}

// TrimGroupEnd will truncate and remove shifts so that the entire group will
// end at the latest exactly t.
func TrimGroupEnd(grp FixedShiftGroup, t time.Time) FixedShiftGroup {
	if !grp.Start.Before(t) {
		// if it doesn't start before t, delete
		return FixedShiftGroup{}
	}
	if !grp.End.After(t) {
		// if it doesn't end after t, no changes
		return grp
	}

	grp.End = t

	newShifts := make([]FixedShift, 0, len(grp.Shifts))
	for _, shift := range grp.Shifts {
		if !shift.Start.Before(t) {
			continue
		}
		if shift.End.After(t) {
			shift.End = t
		}
		newShifts = append(newShifts, shift)
	}
	grp.Shifts = newShifts

	return grp
}

// TrimGroupStart will truncate and remove shifts so that the entire group
// will start at the earliest exactly t.
func TrimGroupStart(grp FixedShiftGroup, t time.Time) FixedShiftGroup {
	if !grp.Start.Before(t) {
		// if it doesn't start before t, no changes
		return grp
	}
	if !grp.End.After(t) {
		// if it doesn't end after t, delete
		return FixedShiftGroup{}
	}

	grp.Start = t

	newShifts := make([]FixedShift, 0, len(grp.Shifts))
	for _, shift := range grp.Shifts {
		if !shift.End.After(t) {
			continue
		}
		if shift.Start.Before(t) {
			shift.Start = t
		}
		newShifts = append(newShifts, shift)
	}
	grp.Shifts = newShifts

	return grp
}

// MergeGroups will sort and merge groups and contained shifts
//
// The output is guaranteed to be in-order and with no overlapping start/end times.
func MergeGroups(groups []FixedShiftGroup) []FixedShiftGroup {
	if len(groups) == 0 {
		return groups
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].Start.Before(groups[j].Start) })
	result := groups[:1]
	for _, g := range groups[1:] {
		l := len(result) - 1
		if g.Start.After(result[l].End) {
			result = append(result, g)
			continue
		}
		result[l].End = maxTime(result[l].End, g.End)
		result[l].Shifts = append(result[l].Shifts, g.Shifts...)
	}

	for i, g := range result {
		g.Shifts = clampShiftTimes(g.Start, g.End, g.Shifts)
		result[i].Shifts = mergeShifts(g.Shifts)
	}

	return result
}

func setFixedShifts(groups []FixedShiftGroup, start, end time.Time, shifts []FixedShift) []FixedShiftGroup {
	groups = deleteFixedShifts(groups, start, end)
	groups = append(groups, FixedShiftGroup{Start: start, End: end, Shifts: shifts})
	return MergeGroups(groups)
}

// deleteFixedShifts will cut groups and shifts out between start and end time.
func deleteFixedShifts(groups []FixedShiftGroup, start, end time.Time) []FixedShiftGroup {
	if !end.After(start) {
		return groups
	}
	result := make([]FixedShiftGroup, 0, len(groups))
	for _, grp := range groups {
		before := TrimGroupEnd(grp, start)
		after := TrimGroupStart(grp, end)
		if !before.Start.IsZero() {
			result = append(result, before)
		}
		if !after.Start.IsZero() {
			result = append(result, after)
		}
	}

	return result
}
