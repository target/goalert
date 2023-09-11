// These tests exercise the requirements specified by ISO:
// https://www.loc.gov/standards/datetime/iso-tc154-wg5_n0038_iso_wd_8601-1_2016-02-16.pdf

package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestISODuration_String(t *testing.T) {
	check := func(exp string, dur ISODuration) {
		t.Helper()

		assert.Equal(t, exp, dur.String())
	}

	check("P1Y", ISODuration{YearPart: 1})
	check("P1Y4M", ISODuration{YearPart: 1, MonthPart: 4})
	check("P1D", ISODuration{DayPart: 1})
	check("PT1H", ISODuration{HourPart: 1})
	check("P1YT0.1S", ISODuration{YearPart: 1, SecondPart: 0.1})

	check("P1Y2M3W4DT5H6M7S", ISODuration{
		YearPart:  1,
		MonthPart: 2,
		WeekPart:  3,
		DayPart:   4,

		HourPart:   5,
		MinutePart: 6,
		SecondPart: 7,
	})
	check("P1Y15D", ISODuration{YearPart: 1, DayPart: 15})
	check("P0D", ISODuration{}) // must contain at least one element
}

func TestParseISODuration(t *testing.T) {

	check := func(desc string, iso string, exp ISODuration) {
		t.Helper()

		res, err := ParseISODuration(iso)
		require.NoError(t, err, desc)
		assert.Equal(t, res, exp, desc)
	}

	check("year only", "P12345Y", ISODuration{
		YearPart: 12345,
	})

	check("one month", "P1M", ISODuration{
		MonthPart: 1,
	})

	check("one minute", "PT1M", ISODuration{
		MinutePart: 1,
	})

	check("one month and 1 minute", "P1MT1M", ISODuration{
		MonthPart:  1,
		MinutePart: 1,
	})

	check("two days with leading zeros", "P0002D", ISODuration{
		// If a time element in a defined representation has a defined length, then leading zeros shall be used as required
		DayPart: 2,
	})

	check("mixed", "P3Y6M14DT12H30M5S", ISODuration{
		YearPart:   3,
		MonthPart:  6,
		DayPart:    14,
		HourPart:   12,
		MinutePart: 30,
		SecondPart: 5,
	})

	check("mixed with week", "P3Y6M2W14DT12H30M5S", ISODuration{
		YearPart:   3,
		MonthPart:  6,
		WeekPart:   2,
		DayPart:    14,
		HourPart:   12,
		MinutePart: 30,
		SecondPart: 5,
	})

	check("time without seconds", "PT1H22M", ISODuration{
		// The lowest order components may be omitted to represent duration with reduced accuracy.
		HourPart:   1,
		MinutePart: 22,
	})

	check("time without minutes", "PT1H22S", ISODuration{
		HourPart:   1,
		SecondPart: 22,
	})

	check("date only", "P1997Y11M26D", ISODuration{
		// The designator [T] shall be absent if all of the time components are absent.
		YearPart:  1997,
		MonthPart: 11,
		DayPart:   26,
	})

	check("week only", "P12W", ISODuration{
		WeekPart: 12,
	})

	check("fractional seconds", "PT0.1S", ISODuration{
		SecondPart: 0.1,
	})

	check("fractional seconds with comma", "PT0,1S", ISODuration{
		// comma [,] is preferred over full stop [.]
		SecondPart: 0.1,
	})

	check("one and a half seconds", "PT1,5S", ISODuration{
		SecondPart: 1.5,
	})

	check("full fractional", "P23Y0M2W012DT1H1M0123.0522S", ISODuration{
		YearPart: 23,
		WeekPart: 2,
		DayPart:  12,

		HourPart:   1,
		MinutePart: 1,
		SecondPart: 123.0522,
	})
}

func TestISODurationFromTime(t *testing.T) {
	check := func(desc string, dur time.Duration, exp string) {
		t.Helper()

		isoDur := ISODurationFromTime(dur)
		assert.Equal(t, dur, isoDur.TimePart(), "TimePart(): "+desc)
		assert.Equal(t, exp, isoDur.String(), desc)
	}

	check("1 second", time.Second, "PT1S")
	check("1 minute", time.Minute, "PT1M")
	check("1 hour", time.Hour, "PT1H")
	check("24 hours", 24*time.Hour, "PT24H")

	// fractional
	check("1.5 seconds", 1500*time.Millisecond, "PT1.5S")
	check("1.5 minutes", 90*time.Second, "PT1M30S")
	check("1.5 hours", 90*time.Minute, "PT1H30M")
}

func TestParseISODurationErrors(t *testing.T) {
	check := func(desc string, iso string) {
		t.Helper()

		_, err := ParseISODuration(iso)
		require.Error(t, err, desc)
	}

	check("empty", "")
	check("P only", "P")
	check("T only", "T")
	check("no units", "PT")
	check("Ends with T", "P1Y1M1DT")
	check("junk", "junk")
	check("missing T", "P1H")
	check("mistaken format", "PY3M6D14TH12M30S5")
	check("missing T 2", "P3Y6M14D12H30M5S")
	check("bad date order", "P1M1Y")
	check("bad time order", "PT1M1H")
	check("missing seconds val", "PTS")
	check("multi decimal", "PT1.2.4S")
	check("missing fractional", "PT1.S")
	check("missing integral", "PT,1S")
}
