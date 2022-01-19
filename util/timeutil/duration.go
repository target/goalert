package timeutil

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/pkg/errors"
)

// ISODuration represents an ISO duration string.
// The time components are combined, and the weeks component
// is interpreted as a shorthand for 7 days.
type ISODuration struct {
	Years, Months, Days int
	TimePart            time.Duration
}

var zeroDur ISODuration

// String returns an ISO 8601 duration string.
func (dur ISODuration) String() string {
	if dur == zeroDur {
		return "P0D"
	}

	var b strings.Builder
	b.WriteRune('P')

	if dur.Years > 0 {
		fmt.Fprintf(&b, "%dY", dur.Years)
	}
	if dur.Months > 0 {
		fmt.Fprintf(&b, "%dM", dur.Months)
	}
	if dur.Days/7 > 0 {
		fmt.Fprintf(&b, "%dW", dur.Days/7)
		dur.Days %= 7
	}
	if dur.Days > 0 {
		fmt.Fprintf(&b, "%dD", dur.Days)
	}

	if dur.TimePart == 0 {
		return b.String()
	}

	b.WriteRune('T')

	if dur.TimePart/time.Hour > 0 {
		fmt.Fprintf(&b, "%dH", dur.TimePart/time.Hour)
		dur.TimePart %= time.Hour
	}

	if dur.TimePart/time.Minute > 0 {
		fmt.Fprintf(&b, "%dM", dur.TimePart/time.Minute)
		dur.TimePart %= time.Minute
	}

	if dur.TimePart.Seconds() > 0 {
		fmt.Fprintf(&b, "%gS", dur.TimePart.Seconds())
	}

	return b.String()
}

var re = regexp.MustCompile(`^P\B(\d+Y)?(\d+M)?(\d+W)?(\d+D)?(T\B(\d+H)?(\d+M)?(\d+S)?)?$`)

// ParseISODuration parses the components of an ISO Duration string.
// The time components are accurate and are aggregated into one TimePart.
// The nominal date components cannot be aggregated without accounting for daylight savings time.
// Supported formats are "PnYnMnDTnHnMnS" and "PnW".
// Negative and decimal units are not supported.
func ParseISODuration(s string) (d ISODuration, err error) {
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
			d.Years += digits
		case "M":
			if isTime {
				digits *= 60
			} else {
				d.Months += digits
			}
		case "D":
			d.Days += digits
		case "W":
			d.Days += (digits * 7)
		case "H":
			digits *= 3600
		case "S":
			// ok
		default:
			return d, errors.Errorf("invalid character encountered: %s", string(c))
		}

		if isTime {
			dur, err := time.ParseDuration(strconv.Itoa(digits) + "s")
			if err != nil {
				return d, err
			}

			d.TimePart += dur
		}

		right++
		left = right
	}

	return d, err
}
