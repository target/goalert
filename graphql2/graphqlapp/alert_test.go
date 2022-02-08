package graphqlapp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/util/timeutil"
)

func TestSplitRangeByDurationAlertCounts(t *testing.T) {

	loc, err := time.LoadLocation("America/Chicago")
	require.NoError(t, err)

	check := func(desc string, r timeutil.ISORInterval, alerts []alert.Alert, exp []int) {
		t.Helper()
		actual := []int{}
		for _, val := range splitRangeByDuration(r, alerts) {
			actual = append(actual, val.AlertCount)
		}
		assert.Equal(t, exp, actual)
	}

	// jan is a test fixture of alerts such that for the first 20 days of Jan,
	// Jan 1 has 1 alert at 12am
	// Jan 2 has 2 alerts at 12am and 1am
	// Jan 3 has 3 alerts at 12, 1, and 2 am
	// ...
	jan := []alert.Alert{}
	for day := 0; day < 20; day++ {
		for hour := 0; hour <= day; hour++ {
			jan = append(jan, alert.Alert{
				CreatedAt: time.Date(2000, time.January, day, hour, 0, 0, 0, loc),
			})
		}
	}

	check(
		"empty alerts",
		timeutil.ISORInterval{
			Repeat: 4,
			Start:  time.Date(2000, time.January, 0, 0, 0, 0, 0, loc),
			Period: timeutil.ISODuration{Days: 7},
		},
		[]alert.Alert{},
		[]int{0, 0, 0, 0, 0},
	)

	check(
		"nil alerts",
		timeutil.ISORInterval{
			Repeat: 4,
			Start:  time.Date(2000, time.January, 0, 0, 0, 0, 0, loc),
			Period: timeutil.ISODuration{Days: 7},
		},
		nil,
		[]int{0, 0, 0, 0, 0},
	)

	check(
		"no alerts",
		timeutil.ISORInterval{
			Repeat: 2,
			Start:  time.Date(1999, time.January, 0, 0, 0, 0, 0, loc),
			Period: timeutil.ISODuration{Days: 1},
		},
		jan,
		[]int{0, 0, 0},
	)

	check(
		"Jan 1st",
		timeutil.ISORInterval{
			Repeat: 0,
			Start:  time.Date(2000, time.January, 0, 0, 0, 0, 0, loc),
			Period: timeutil.ISODuration{Days: 1},
		},
		jan,
		[]int{1},
	)

	check(
		"Jan 2nd",
		timeutil.ISORInterval{
			Repeat: 0,
			Start:  time.Date(2000, time.January, 1, 0, 0, 0, 0, loc),
			Period: timeutil.ISODuration{Days: 1},
		},
		jan,
		[]int{2},
	)

	check(
		"Jan 1st and 2nd",
		timeutil.ISORInterval{
			Repeat: 1,
			Start:  time.Date(2000, time.January, 0, 0, 0, 0, 0, loc),
			Period: timeutil.ISODuration{Days: 1},
		},
		jan,
		[]int{1, 2},
	)

	check(
		"Jan 1st thru 15th",
		timeutil.ISORInterval{
			Repeat: 14,
			Start:  time.Date(2000, time.January, 0, 0, 0, 0, 0, loc),
			Period: timeutil.ISODuration{Days: 1},
		},
		jan,
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
	)

	check(
		"Jan 1st thru 16th, 2-day chunks",
		timeutil.ISORInterval{
			Repeat: 7,
			Start:  time.Date(2000, time.January, 0, 0, 0, 0, 0, loc),
			Period: timeutil.ISODuration{Days: 2},
		},
		jan,
		[]int{3, 7, 11, 15, 19, 23, 27, 31},
	)

	check(
		"Jan weekly chunks",
		timeutil.ISORInterval{
			Repeat: 4,
			Start:  time.Date(2000, time.January, 0, 0, 0, 0, 0, loc),
			Period: timeutil.ISODuration{Days: 7},
		},
		jan,
		[]int{28, 77, 105, 0, 0},
	)

}
