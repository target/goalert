package alert

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/util/timeutil"
)

func TestSplitRangeByDurationAlertCounts(t *testing.T) {

	loc, err := time.LoadLocation("America/Chicago")
	require.NoError(t, err)

	jan := []Alert{}
	for day := 0; day < 20; day++ {
		for hour := 0; hour <= day; hour++ {
			jan = append(jan, Alert{
				CreatedAt: time.Date(2000, time.January, day, hour, 0, 0, 0, loc),
			})
		}
	}

	check := func(desc string, since, until time.Time, ISOduration string, exp []int) {
		t.Helper()
		dur, err := timeutil.ParseISODuration(ISOduration)
		require.NoError(t, err)

		var actual []int
		for _, val := range SplitRangeByDuration(since, until, dur, jan) {
			actual = append(actual, val.AlertCount)
		}
		assert.Equal(t, exp, actual)
	}

	check(
		"since == until",
		time.Date(2000, time.January, 0, 0, 0, 0, 0, loc),
		time.Date(2000, time.January, 0, 0, 0, 0, 0, loc),
		"P1D",
		nil,
	)

	check(
		"since before until",
		time.Date(9999, time.January, 0, 0, 0, 0, 0, loc),
		time.Date(2000, time.January, 0, 0, 0, 0, 0, loc),
		"P1D",
		nil,
	)

	check(
		"no alerts",
		time.Date(1999, time.January, 0, 0, 0, 0, 0, loc),
		time.Date(1999, time.January, 2, 3, 4, 5, 6, loc),
		"P1D",
		[]int{0, 0, 0},
	)

	check(
		"Jan 1st",
		time.Date(2000, time.January, 0, 0, 0, 0, 0, loc),
		time.Date(2000, time.January, 1, 0, 0, 0, 0, loc),
		"P1D",
		[]int{1},
	)

	check(
		"Jan 2nt",
		time.Date(2000, time.January, 1, 0, 0, 0, 0, loc),
		time.Date(2000, time.January, 2, 0, 0, 0, 0, loc),
		"P1D",
		[]int{2},
	)

	check(
		"Jan 1st and 2nd",
		time.Date(2000, time.January, 0, 0, 0, 0, 0, loc),
		time.Date(2000, time.January, 2, 0, 0, 0, 0, loc),
		"P1D",
		[]int{1, 2},
	)

	check(
		"Jan 1st thru 15th",
		time.Date(2000, time.January, 0, 0, 0, 0, 0, loc),
		time.Date(2000, time.January, 15, 0, 0, 0, 0, loc),
		"P1D",
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
	)

	check(
		"Jan 1st thru 15th, 2-day chunks",
		time.Date(2000, time.January, 0, 0, 0, 0, 0, loc),
		time.Date(2000, time.January, 15, 0, 0, 0, 0, loc),
		"P2D",
		[]int{3, 7, 11, 15, 19, 23, 27, 15},
	)

}
