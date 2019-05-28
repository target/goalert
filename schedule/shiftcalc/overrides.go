package shiftcalc

import (
	"github.com/target/goalert/override"
	"sort"
	"time"
)

func (d *data) ScheduleFinalShiftsWithOverrides(start, end time.Time) []Shift {
	return finalShiftsWithOverrides(d.ScheduleFinalShifts(start, end), d.userOverrides)
}

func applyOverride(shift Shift, o override.UserOverride) (result []Shift) {
	if shift.UserID != o.RemoveUserID {
		return []Shift{shift}
	}
	if !o.Start.Before(shift.End) {
		return []Shift{shift}
	}
	if !o.End.After(shift.Start) {
		return []Shift{shift}
	}

	if shift.Start.Before(o.Start) {
		// break off first part of shift
		result = append(result, Shift{UserID: shift.UserID, Start: shift.Start, End: o.Start})

		// advance the start of the remaining shift being processed
		shift.Start = o.Start
	}

	// at this point we know that shift start is during override

	// we're "replacing", so the AddUserID is on-call during the override
	end := o.End // end is the end of the override or shift, whichever is sooner
	if end.After(shift.End) {
		end = shift.End
	}
	if o.AddUserID != "" {
		result = append(result, Shift{UserID: o.AddUserID, Start: shift.Start, End: end})
	}

	if end.Before(shift.End) {
		// Original user completes their shift
		result = append(result, Shift{UserID: shift.UserID, Start: end, End: shift.End})
	}

	return result
}

func finalShiftsWithOverrides(final []Shift, userOverrides []override.UserOverride) []Shift {
	withOverrides := make([]Shift, 0, len(final))

	addOverrides := make([]override.UserOverride, 0, len(userOverrides))
	otherOverrides := make([]override.UserOverride, 0, len(userOverrides))
	for _, o := range userOverrides {
		if o.RemoveUserID == "" {
			addOverrides = append(addOverrides, o)
			continue
		}

		otherOverrides = append(otherOverrides, o)
	}

	// Sort overrides by start time so that as we progress the .Start of the shift forward
	// they don't get skipped over.
	sort.Slice(otherOverrides, func(i, j int) bool { return otherOverrides[i].Start.Before(otherOverrides[j].Start) })

	a := make([]Shift, 0, len(final)*2)
	b := make([]Shift, 0, len(final)*2)
	a = append(a, final...)
	for _, o := range otherOverrides {
		for _, shift := range a {
			b = append(b, applyOverride(shift, o)...)
		}
		a, b = b, a
		b = b[:0]
	}
	withOverrides = append(withOverrides, a...)

	for _, o := range addOverrides {
		withOverrides = append(withOverrides, Shift{
			Start:  o.Start,
			End:    o.End,
			UserID: o.AddUserID,
		})
	}

	return mergeShiftsByTarget(withOverrides)
}
