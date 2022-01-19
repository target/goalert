// These tests exercise the requirements specified by ISO:
// https://www.loc.gov/standards/datetime/iso-tc154-wg5_n0038_iso_wd_8601-1_2016-02-16.pdf

package timeutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseISODuration(t *testing.T) {

	check := func(desc string, iso string, exp ISODuration) {
		t.Helper()

		res, err := ParseISODuration(iso)
		require.NoError(t, err, desc)
		assert.Equal(t, res, exp, desc)
	}

	check("year only", "P12345Y", ISODuration{
		years: 12345,
	})

	check("one month", "P1M", ISODuration{
		months: 1,
	})

	check("one minute", "PT1M", ISODuration{
		seconds: 60,
	})

	check("one month and 1 minute", "P1MT1M", ISODuration{
		months:  1,
		seconds: 60,
	})

	check("two days with leading zeros", "P0002D", ISODuration{
		// If a time element in a defined representation has a defined length, then leading zeros shall be used as required
		days: 2,
	})

	check("mixed", "P3Y6M14DT12H30M5S", ISODuration{
		years:   3,
		months:  6,
		days:    14,
		seconds: 12*3600 + 30*60 + 5,
	})

	check("mixed with week", "P3Y6M2W14DT12H30M5S", ISODuration{
		years:   3,
		months:  6,
		days:    2*7 + 14,
		seconds: 12*3600 + 30*60 + 5,
	})

	check("time without seconds", "PT1H22M", ISODuration{
		// The lowest order components may be omitted to represent duration with reduced accuracy.
		seconds: 3600 + 22*60,
	})

	check("time without minutes", "PT1H22S", ISODuration{
		seconds: 3600 + 22,
	})

	check("date only", "P1997Y11M26D", ISODuration{
		// The designator [T] shall be absent if all of the time components are absent.
		years:  1997,
		months: 11,
		days:   26,
	})

	check("week only", "P12W", ISODuration{
		days: 12 * 7,
	})

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
	check("Ends with T", "P1Y1M1DT")
	check("junk", "junk")
	check("missing T", "P1H")
	check("mistaken format", "PY3M6D14TH12M30S5")
	check("missing T 2", "P3Y6M14D12H30M5S")
	check("bad date order", "P1M1Y")
	check("bad time order", "PT1M1H")
}
