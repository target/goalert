package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNextZoneTransition(t *testing.T) {
	load := func(name string) *time.Location {
		t.Helper()
		loc, err := time.LoadLocation(name)
		require.NoError(t, err)
		return loc
	}
	chi := load("America/Chicago")
	stgo := load("America/Santiago")
	lh := load("Australia/Lord_Howe")

	utc := func(mo time.Month, d, h, m, s int) time.Time {
		return time.Date(2026, mo, d, h, m, s, 0, time.UTC)
	}

	// Chicago: 2026-03-08 02:00 CST -> 03:00 CDT is 08:00Z; 2026-11-01 02:00 CDT -> 01:00 CST is 07:00Z.
	chiSpring := utc(time.March, 8, 8, 0, 0)
	chiFall := utc(time.November, 1, 7, 0, 0)
	// Santiago: 2026-09-06 00:00 -04 -> 01:00 -03 is 04:00Z (a midnight transition).
	stgoSpring := utc(time.September, 6, 4, 0, 0)
	// Lord Howe: 2026-10-04 02:00 +10:30 -> 02:30 +11:00 is 2026-10-03 15:30Z (a 30-minute shift).
	lhSpring := utc(time.October, 3, 15, 30, 0)

	cases := []struct {
		name  string
		t     time.Time
		until time.Time
		exp   time.Time
		ok    bool
	}{
		{
			name:  "chicago spring forward found in 24h window",
			t:     time.Date(2026, 3, 8, 0, 0, 0, 0, chi),
			until: time.Date(2026, 3, 8, 0, 0, 0, 0, chi).Add(24 * time.Hour),
			exp:   chiSpring,
			ok:    true,
		},
		{
			name:  "chicago fall back found in 24h window",
			t:     time.Date(2026, 11, 1, 0, 0, 0, 0, chi),
			until: time.Date(2026, 11, 1, 0, 0, 0, 0, chi).Add(24 * time.Hour),
			exp:   chiFall,
			ok:    true,
		},
		{
			name:  "no transition",
			t:     time.Date(2026, 10, 21, 0, 0, 0, 0, chi),
			until: time.Date(2026, 10, 21, 0, 0, 0, 0, chi).Add(24 * time.Hour),
		},
		{
			name:  "until before the transition",
			t:     time.Date(2026, 3, 8, 0, 0, 0, 0, chi),
			until: time.Date(2026, 3, 8, 1, 30, 0, 0, chi),
		},
		{
			name:  "until exactly at the transition",
			t:     time.Date(2026, 3, 8, 0, 0, 0, 0, chi),
			until: chiSpring,
			exp:   chiSpring,
			ok:    true,
		},
		{
			name:  "t with seconds yields the minute-aligned instant",
			t:     time.Date(2026, 3, 8, 1, 59, 30, 0, chi),
			until: time.Date(2026, 3, 8, 1, 59, 30, 0, chi).Add(time.Hour),
			exp:   chiSpring,
			ok:    true,
		},
		{
			name:  "t and until within the same minute around the transition",
			t:     chiSpring.Add(-10 * time.Second).In(chi),
			until: chiSpring.Add(10 * time.Second),
			exp:   chiSpring,
			ok:    true,
		},
		{
			name:  "until in a different location",
			t:     time.Date(2026, 3, 8, 0, 0, 0, 0, chi),
			until: chiSpring.Add(time.Hour), // UTC
			exp:   chiSpring,
			ok:    true,
		},
		{
			name:  "window starting exactly at a transition",
			t:     chiSpring.In(chi),
			until: chiSpring.Add(24 * time.Hour),
		},
		{
			name:  "until not after t",
			t:     time.Date(2026, 3, 8, 0, 0, 0, 0, chi),
			until: time.Date(2026, 3, 8, 0, 0, 0, 0, chi),
		},
		{
			name:  "until before t",
			t:     time.Date(2026, 3, 8, 0, 0, 0, 0, chi),
			until: time.Date(2026, 3, 7, 0, 0, 0, 0, chi),
		},
		{
			name:  "santiago midnight transition",
			t:     time.Date(2026, 9, 5, 20, 0, 0, 0, stgo),
			until: time.Date(2026, 9, 5, 20, 0, 0, 0, stgo).Add(12 * time.Hour),
			exp:   stgoSpring,
			ok:    true,
		},
		{
			name:  "santiago window ending before midnight transition",
			t:     time.Date(2026, 9, 5, 20, 0, 0, 0, stgo),
			until: time.Date(2026, 9, 5, 20, 0, 0, 0, stgo).Add(time.Hour),
		},
		{
			name:  "lord howe 30-minute shift",
			t:     time.Date(2026, 10, 4, 0, 0, 0, 0, lh),
			until: time.Date(2026, 10, 4, 0, 0, 0, 0, lh).Add(6 * time.Hour),
			exp:   lhSpring,
			ok:    true,
		},
		{
			name:  "utc never transitions",
			t:     utc(time.March, 8, 0, 0, 0),
			until: utc(time.March, 9, 0, 0, 0),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := NextZoneTransition(c.t, c.until)
			assert.Equal(t, c.ok, ok, "found")
			if !c.ok {
				assert.True(t, got.IsZero(), "expected zero time, got %s", got)
				return
			}

			assert.Equal(t, c.exp.UTC().Format(time.RFC3339Nano), got.UTC().Format(time.RFC3339Nano))
			assert.True(t, got.After(c.t), "result must be after t")
			assert.False(t, got.After(c.until), "result must not be after until")

			_, oldOffset := c.t.Zone()
			_, newOffset := got.In(c.t.Location()).Zone()
			assert.NotEqual(t, oldOffset, newOffset, "offset must differ at the transition")
			_, prevOffset := got.Add(-time.Second).In(c.t.Location()).Zone()
			assert.Equal(t, oldOffset, prevOffset, "offset must be unchanged just before the transition")
		})
	}
}
