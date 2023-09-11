package timeutil_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/util/timeutil"
)

func TestParseISORInterval(t *testing.T) {
	check := func(input string, exp timeutil.ISORInterval, expEnd time.Time) {
		t.Helper()

		res, err := timeutil.ParseISORInterval(input)
		require.NoError(t, err)
		assert.Equal(t, exp, res)
		assert.Equal(t, expEnd, res.End())
	}

	check("R0/2022-01-31T00:00:00Z/2023-01-31T00:00:00Z", timeutil.ISORInterval{
		Repeat: 0,
		Start:  time.Date(2022, 1, 31, 0, 0, 0, 0, time.UTC),
		Period: timeutil.ISODuration{HourPart: 24 * 365},
	}, time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC))

	check("R1/2022-01-31T00:00:00Z/2024-01-31T00:00:00Z", timeutil.ISORInterval{
		Repeat: 1,
		Start:  time.Date(2022, 1, 31, 0, 0, 0, 0, time.UTC),
		Period: timeutil.ISODuration{HourPart: 24 * 365},
	}, time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC))

	check("R0/2022-01-31T00:00:00Z/P1Y", timeutil.ISORInterval{
		Repeat: 0,
		Start:  time.Date(2022, 1, 31, 0, 0, 0, 0, time.UTC),
		Period: timeutil.ISODuration{YearPart: 1},
	}, time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC))

	check("R1/2022-01-31T00:00:00Z/P1Y", timeutil.ISORInterval{
		Repeat: 1,
		Start:  time.Date(2022, 1, 31, 0, 0, 0, 0, time.UTC),
		Period: timeutil.ISODuration{YearPart: 1},
	}, time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC))

	check("R0/P1Y/2023-01-31T00:00:00Z", timeutil.ISORInterval{
		Repeat: 0,
		Start:  time.Date(2022, 1, 31, 0, 0, 0, 0, time.UTC),
		Period: timeutil.ISODuration{YearPart: 1},
	}, time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC))

	check("R1/P1Y/2024-01-31T00:00:00Z", timeutil.ISORInterval{
		Repeat: 1,
		Start:  time.Date(2022, 1, 31, 0, 0, 0, 0, time.UTC),
		Period: timeutil.ISODuration{YearPart: 1},
	}, time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC))

}

func TestISORInterval_String(t *testing.T) {
	check := func(exp string, r timeutil.ISORInterval) {
		t.Helper()

		assert.Equal(t, exp, r.String())
	}

	check("R0/2022-01-31T00:00:00Z/P1Y", timeutil.ISORInterval{
		Repeat: 0,
		Start:  time.Date(2022, 1, 31, 0, 0, 0, 0, time.UTC),
		Period: timeutil.ISODuration{YearPart: 1},
	})
	check("R1/2022-01-31T00:00:00Z/P1Y", timeutil.ISORInterval{
		Repeat: 1,
		Start:  time.Date(2022, 1, 31, 0, 0, 0, 0, time.UTC),
		Period: timeutil.ISODuration{YearPart: 1},
	})

	check("R0/2022-01-31T00:00:00Z/2022-01-31T01:00:00Z", timeutil.ISORInterval{
		Repeat: 0,
		Start:  time.Date(2022, 1, 31, 0, 0, 0, 0, time.UTC),
		Period: timeutil.ISODuration{HourPart: 1},
	})
	check("R1/2022-01-31T00:00:00Z/2022-01-31T02:00:00Z", timeutil.ISORInterval{
		Repeat: 1,
		Start:  time.Date(2022, 1, 31, 0, 0, 0, 0, time.UTC),
		Period: timeutil.ISODuration{HourPart: 1},
	})
}
