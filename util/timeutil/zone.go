package timeutil

import (
	"sort"
	"time"
)

// NextZoneTransition returns the first instant in (t, until] at which the UTC
// offset of t.Location() differs from the offset in effect at t, along with
// true. If the offset at until is the same as the offset at t, the zero time
// and false are returned.
//
// The transition instant is located by binary search at minute resolution
// (the same approach as IsDST), so the returned value is aligned to the
// minute grid of t (i.e., t.Truncate(time.Minute) plus whole minutes). All
// modern zone transitions occur on whole-minute boundaries, so in practice
// the exact transition instant is returned.
//
// Windows are expected to be short (a day or less) so that at most one
// transition falls within them: only the offsets at the two endpoints are
// compared, so a window containing two transitions that return to the
// original offset reports no transition at all.
func NextZoneTransition(t, until time.Time) (time.Time, bool) {
	if !until.After(t) {
		return time.Time{}, false
	}

	loc := t.Location()
	until = until.In(loc)
	_, oldOffset := t.Zone()
	_, newOffset := until.Zone()
	if oldOffset == newOffset {
		return time.Time{}, false
	}

	// Search on the minute grid starting from t truncated to the minute, so
	// that a t with a seconds component still yields the minute-aligned
	// transition instant. The first candidate (k=1) is always after t.
	base := t.Truncate(time.Minute).In(loc)
	n := int(until.Sub(base) / time.Minute)
	if n < 1 {
		// Sub-minute window; the transition is somewhere in (t, until].
		return until, true
	}

	k := sort.Search(n, func(i int) bool {
		_, off := base.Add(time.Duration(i+1) * time.Minute).Zone()
		return off == newOffset
	})
	if k == n {
		// No whole-minute candidate <= until carries the new offset, so the
		// transition lies in the final partial minute before until.
		return until, true
	}

	return base.Add(time.Duration(k+1) * time.Minute), true
}
