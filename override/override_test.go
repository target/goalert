package override

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/assignment"
)

func TestUserOverride_Normalize(t *testing.T) {
	const (
		userID  = "0b7f2b3e-4c5d-4e6f-8a9b-0c1d2e3f4a5b"
		schedID = "1c8e3c4f-5d6e-4f70-9bac-1d2e3f4a5b6c"
	)
	tgt := assignment.ScheduleTarget(schedID)

	t.Run("truncates start and end to the minute", func(t *testing.T) {
		in := UserOverride{
			AddUserID: userID,
			Target:    tgt,
			Start:     time.Date(2000, 1, 1, 10, 0, 30, 500, time.UTC),
			End:       time.Date(2000, 1, 1, 12, 30, 59, 999999999, time.UTC),
		}

		res, err := in.Normalize()
		require.NoError(t, err)

		assert.Equal(t, time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC), res.Start, "Start")
		assert.Equal(t, time.Date(2000, 1, 1, 12, 30, 0, 0, time.UTC), res.End, "End")
		assert.Equal(t, 0, res.Start.Second(), "Start seconds dropped")
		assert.Equal(t, 0, res.Start.Nanosecond(), "Start nanoseconds dropped")
		assert.Equal(t, 0, res.End.Second(), "End seconds dropped")
		assert.Equal(t, 0, res.End.Nanosecond(), "End nanoseconds dropped")

		// Normalize returns a copy; the input must be unchanged.
		assert.Equal(t, time.Date(2000, 1, 1, 10, 0, 30, 500, time.UTC), in.Start, "input Start untouched")
		assert.Equal(t, time.Date(2000, 1, 1, 12, 30, 59, 999999999, time.UTC), in.End, "input End untouched")
	})

	t.Run("minute-aligned times are unchanged", func(t *testing.T) {
		start := time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC)
		end := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
		res, err := UserOverride{
			RemoveUserID: userID,
			Target:       tgt,
			Start:        start,
			End:          end,
		}.Normalize()
		require.NoError(t, err)

		assert.True(t, res.Start.Equal(start), "Start")
		assert.True(t, res.End.Equal(end), "End")
	})

	t.Run("preserves timezone", func(t *testing.T) {
		loc := time.FixedZone("UTC-6", -6*60*60)
		res, err := UserOverride{
			AddUserID: userID,
			Target:    tgt,
			Start:     time.Date(2000, 1, 1, 10, 15, 45, 0, loc),
			End:       time.Date(2000, 1, 1, 18, 45, 15, 0, loc),
		}.Normalize()
		require.NoError(t, err)

		assert.Equal(t, loc, res.Start.Location(), "Start location")
		assert.Equal(t, loc, res.End.Location(), "End location")
		assert.Equal(t, time.Date(2000, 1, 1, 10, 15, 0, 0, loc), res.Start, "Start")
		assert.Equal(t, time.Date(2000, 1, 1, 18, 45, 0, 0, loc), res.End, "End")
	})

	t.Run("start equal to end after truncation is rejected", func(t *testing.T) {
		// Distinct instants that collapse to the same minute.
		_, err := UserOverride{
			AddUserID: userID,
			Target:    tgt,
			Start:     time.Date(2000, 1, 1, 10, 0, 1, 0, time.UTC),
			End:       time.Date(2000, 1, 1, 10, 0, 59, 0, time.UTC),
		}.Normalize()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "End")
	})

	t.Run("start equal to end is rejected", func(t *testing.T) {
		ts := time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC)
		_, err := UserOverride{
			AddUserID: userID,
			Target:    tgt,
			Start:     ts,
			End:       ts,
		}.Normalize()
		require.Error(t, err)
	})

	t.Run("end before start is rejected", func(t *testing.T) {
		_, err := UserOverride{
			AddUserID: userID,
			Target:    tgt,
			Start:     time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
			End:       time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC),
		}.Normalize()
		require.Error(t, err)
	})

	t.Run("sub-minute span across a minute boundary is kept", func(t *testing.T) {
		// Start truncates down to 10:00, End stays at 10:01: still ordered, still valid.
		res, err := UserOverride{
			AddUserID: userID,
			Target:    tgt,
			Start:     time.Date(2000, 1, 1, 10, 0, 59, 0, time.UTC),
			End:       time.Date(2000, 1, 1, 10, 1, 0, 0, time.UTC),
		}.Normalize()
		require.NoError(t, err)

		assert.Equal(t, time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC), res.Start, "Start")
		assert.Equal(t, time.Date(2000, 1, 1, 10, 1, 0, 0, time.UTC), res.End, "End")
	})

	t.Run("requires add and/or remove user", func(t *testing.T) {
		_, err := UserOverride{
			Target: tgt,
			Start:  time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC),
			End:    time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		}.Normalize()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "UserID")
	})
}
