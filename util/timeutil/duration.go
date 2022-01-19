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
	var nextDigits []rune
	var isTime bool

	re := regexp.MustCompile(`^P\B(\d+Y)?(\d+M)?(\d+W)?(\d+D)?(T\B(\d+H)?(\d+M)?(\d+S)?)?$`)

	if !re.MatchString(s) {
		return d, errors.Errorf(`invalid format: %s must be an ISO Duration`, s)
	}

	for _, c := range s[1:] {
		if unicode.IsDigit(c) {
			nextDigits = append(nextDigits, c)
			continue
		}

		switch string(c) {
		case "Y":
			d.years += runesToInt(nextDigits)
		case "M":
			if isTime {
				// minutes
				d.seconds += (runesToInt(nextDigits) * 60)
			} else {
				d.months += runesToInt(nextDigits)
			}
		case "D":
			d.days += runesToInt(nextDigits)
		case "W":
			d.days += (runesToInt(nextDigits) * 7)
		case "H":
			d.seconds += (runesToInt(nextDigits) * 60 * 60)
		case "S":
			d.seconds += runesToInt(nextDigits)
		case "T":
			isTime = true
		default:
			return d, errors.Errorf("invalid character encountered: %s", string(c))
		}

		nextDigits = nextDigits[:0]

	}

	return d, err
}

func runesToInt(r []rune) int {
	res, _ := strconv.Atoi(string(r))
	return res
}
