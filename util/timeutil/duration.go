package timeutil

import (
	"regexp"
	"strconv"
	"unicode"

	"github.com/pkg/errors"
)

type ISODuration struct {
	years, months, days, seconds int
}

// ParseISODuration parses the components of an ISO Duration string into years, months, days, and seconds.
// The time components are aggregated into seconds as "the base unit for expressing duration".
// Nominal date components cannot be aggregated without accounting for daylight savings time, except weeks,
// which are interpreted as a shorthand for 7 days.
// Supported formats are "PnYnMnDTnHnMnS" and "PnW".
// Negative and decimal units are not supported.
func ParseISODuration(s string) (d ISODuration, err error) {
	re := regexp.MustCompile(`^P\B(\d+Y)?(\d+M)?(\d+W)?(\d+D)?(T\B(\d+H)?(\d+M)?(\d+S)?)?$`)
	if !re.MatchString(s) {
		return d, errors.Errorf(`invalid format: %s must be an ISO Duration`, s)
	}

	left, right := 1, 1 // sliding window
	isTime := false

	for _, c := range s[1:] {
		if unicode.IsDigit(c) {
			right++
			continue
		}

		if string(c) == "T" {
			isTime = true
			right++
			left = right
			continue
		}

		digits, err := strconv.Atoi(s[left:right])
		if err != nil {
			return d, err
		}

		switch string(c) {
		case "Y":
			d.years += digits
		case "M":
			if isTime {
				// minutes
				d.seconds += (digits * 60)
			} else {
				d.months += digits
			}
		case "D":
			d.days += digits
		case "W":
			d.days += (digits * 7)
		case "H":
			d.seconds += (digits * 60 * 60)
		case "S":
			d.seconds += digits
		default:
			return d, errors.Errorf("invalid character encountered: %s", string(c))
		}

		right++
		left = right
	}

	return d, err
}
