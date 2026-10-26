package schedulemanager

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/util/timeutil"
)

func TestNextOnCallNotification(t *testing.T) {
	now := time.Date(2021, 7, 7, 11, 0, 0, 0, time.UTC)

	check := func(desc string, _time *timeutil.Clock, _filter *timeutil.WeekdayFilter, exp time.Time) {
		t.Helper()
		t.Run(desc, func(t *testing.T) {
			res := nextOnCallNotification(now, schedule.OnCallNotificationRule{
				Time:          _time,
				WeekdayFilter: _filter,
			})
			if exp.IsZero() {
				assert.Nil(t, res)
				return
			}

			require.NotNil(t, res)

			assert.Equal(t, exp.In(time.UTC).String(), res.In(time.UTC).String())
		})
	}

	check("no time of day", nil, nil, time.Time{})

	clock := timeutil.NewClock(4, 0)
	check("no filter", &clock, nil, time.Date(2021, 7, 8, 4, 0, 0, 0, time.UTC))

	clock = timeutil.NewClock(4, 0)
	filter := timeutil.EveryDay()
	check("always filter", &clock, &filter, time.Date(2021, 7, 8, 4, 0, 0, 0, time.UTC))

	clock = timeutil.NewClock(11, 5)
	filter = timeutil.EveryDay()
	check("always filter, later today", &clock, &filter, time.Date(2021, 7, 7, 11, 5, 0, 0, time.UTC))

	filter = timeutil.WeekdayFilter{}
	check("never filter", &clock, &filter, time.Time{})

	clock = timeutil.NewClock(4, 0)
	filter = timeutil.WeekdayFilter{0, 0, 0, 0, 1, 0, 0}
	check("next day", &clock, &filter, time.Date(2021, 7, 8, 4, 0, 0, 0, time.UTC))

	clock = timeutil.NewClock(11, 5)
	filter = timeutil.WeekdayFilter{0, 0, 0, 1, 0, 0, 0}
	check("later today", &clock, &filter, time.Date(2021, 7, 7, 11, 5, 0, 0, time.UTC))

	clock = timeutil.NewClock(4, 0)
	filter = timeutil.WeekdayFilter{0, 0, 0, 1, 0, 0, 0}
	check("prev day", &clock, &filter, time.Date(2021, 7, 14, 4, 0, 0, 0, time.UTC))

	clock = timeutil.NewClock(11, 0)
	filter = timeutil.WeekdayFilter{0, 0, 0, 1, 0, 0, 0}
	check("now", &clock, &filter, time.Date(2021, 7, 14, 11, 0, 0, 0, time.UTC))

}

func TestNextOnCallNotification_MidnightDST(t *testing.T) {
	// DST starts at midnight in Santiago on Sunday 2026-09-06 (00:00 -> 01:00),
	// so local midnight does not exist on that day.
	loc, err := time.LoadLocation("America/Santiago")
	require.NoError(t, err)

	check := func(desc string, now time.Time, clock timeutil.Clock, filter *timeutil.WeekdayFilter, exp string) {
		t.Helper()
		t.Run(desc, func(t *testing.T) {
			res := nextOnCallNotification(now, schedule.OnCallNotificationRule{
				Time:          &clock,
				WeekdayFilter: filter,
			})
			require.NotNil(t, res)
			assert.Equal(t, exp, res.String())
			assert.True(t, res.After(now), "should be in the future")
		})
	}

	sat := time.Date(2026, time.September, 5, 10, 0, 0, 0, loc)
	sun := time.Date(2026, time.September, 6, 1, 0, 0, 0, loc)

	check("no filter, later today", sat, timeutil.NewClock(11, 0), nil, "2026-09-05 11:00:00 -0400 -04")
	check("no filter, next day", sat, timeutil.NewClock(9, 0), nil, "2026-09-06 09:00:00 -0300 -03")
	check("no filter, time does not exist", sat, timeutil.NewClock(0, 30), nil, "2026-09-06 01:00:00 -0300 -03")
	check("no filter, on the gap day", sun, timeutil.NewClock(9, 0), nil, "2026-09-06 09:00:00 -0300 -03")
	check("no filter, on the gap day, next day", sun, timeutil.NewClock(0, 30), nil, "2026-09-07 00:30:00 -0300 -03")

	sunOnly := timeutil.WeekdayFilter{1, 0, 0, 0, 0, 0, 0}
	check("filter, next day", sat, timeutil.NewClock(9, 0), &sunOnly, "2026-09-06 09:00:00 -0300 -03")
	check("filter, time does not exist", sat, timeutil.NewClock(0, 30), &sunOnly, "2026-09-06 01:00:00 -0300 -03")
	check("filter, next week", sun, timeutil.NewClock(0, 30), &sunOnly, "2026-09-13 00:30:00 -0300 -03")
}
