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
