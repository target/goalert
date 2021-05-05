package schedule

import (
	"sort"
	"time"
)

// TemporarySchedule represents a timespan containing static pre-defined shifts of on-call users.
type TemporarySchedule struct {
	Start, End time.Time
	Shifts     []FixedShift
}

// TrimEnd will truncate and remove shifts so that the entire TemporarySchedule will
// end at the latest exactly t.
func (temp TemporarySchedule) TrimEnd(t time.Time) TemporarySchedule {
	if !temp.Start.Before(t) {
		// if it doesn't start before t, delete
		return TemporarySchedule{}
	}
	if !temp.End.After(t) {
		// if it doesn't end after t, no changes
		return temp
	}

	temp.End = t

	newShifts := make([]FixedShift, 0, len(temp.Shifts))
	for _, shift := range temp.Shifts {
		if !shift.Start.Before(t) {
			continue
		}
		if shift.End.After(t) {
			shift.End = t
		}
		newShifts = append(newShifts, shift)
	}
	temp.Shifts = newShifts

	return temp
}

// TrimStart will truncate and remove shifts so that the entire TemporarySchedule
// will start at the earliest exactly t.
func (temp TemporarySchedule) TrimStart(t time.Time) TemporarySchedule {
	if !temp.Start.Before(t) {
		// if it doesn't start before t, no changes
		return temp
	}
	if !temp.End.After(t) {
		// if it doesn't end after t, delete
		return TemporarySchedule{}
	}

	temp.Start = t

	newShifts := make([]FixedShift, 0, len(temp.Shifts))
	for _, shift := range temp.Shifts {
		if !shift.End.After(t) {
			continue
		}
		if shift.Start.Before(t) {
			shift.Start = t
		}
		newShifts = append(newShifts, shift)
	}
	temp.Shifts = newShifts

	return temp
}

// MergeTemporarySchedules will sort and merge TemporarySchedules and contained shifts
//
// The output is guaranteed to be in-order and with no overlapping start/end times.
func MergeTemporarySchedules(tempScheds []TemporarySchedule) []TemporarySchedule {
	if len(tempScheds) == 0 {
		return tempScheds
	}
	sort.Slice(tempScheds, func(i, j int) bool { return tempScheds[i].Start.Before(tempScheds[j].Start) })
	result := tempScheds[:1]
	for _, g := range tempScheds[1:] {
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

func setFixedShifts(tempScheds []TemporarySchedule, newSched TemporarySchedule) []TemporarySchedule {
	tempScheds = deleteFixedShifts(tempScheds, newSched.Start, newSched.End)
	tempScheds = append(tempScheds, newSched)
	return MergeTemporarySchedules(tempScheds)
}

// deleteFixedShifts will cut TemporarySchedules and shifts out between start and end time.
func deleteFixedShifts(tempScheds []TemporarySchedule, start, end time.Time) []TemporarySchedule {
	if !end.After(start) {
		return tempScheds
	}
	result := make([]TemporarySchedule, 0, len(tempScheds))
	for _, temp := range tempScheds {
		before := temp.TrimEnd(start)
		after := temp.TrimStart(end)
		if !before.Start.IsZero() {
			result = append(result, before)
		}
		if !after.Start.IsZero() {
			result = append(result, after)
		}
	}

	return result
}
