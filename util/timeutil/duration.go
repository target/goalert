package timeutil

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
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

var re = regexp.MustCompile(`^P\B((?P<year>\d+)Y)?((?P<month>\d+)M)?((?P<week>\d+)W)?((?P<day>\d+)D)?(T\B((?P<hour>\d+)H)?((?P<minute>\d+)M)?((?P<second>\d*[.,]?\d+)S)?)?$`)

// ParseISODuration parses the components of an ISO Duration string.
// The time components are accurate and are aggregated into one TimePart.
// The nominal date components cannot be aggregated without accounting for daylight savings time.
// Supported formats are "PnYnMnDTnHnMnS" and "PnW".
// Negative values are not supported. Fractional values are only supported for seconds.
func ParseISODuration(s string) (d ISODuration, err error) {
	if !re.MatchString(s) {
		return zeroDur, fmt.Errorf("invalid ISO Duration format: %s", s)
	}

	matches := re.FindStringSubmatch(s)

	for i, name := range re.SubexpNames() {
		m := matches[i]
		if i == 0 || name == "" || m == "" {
			continue
		}

		switch name {
		case "year":
			val, err := strconv.Atoi(m)
			if err != nil {
				return zeroDur, err
			}
			d.Years += val
		case "month":
			val, err := strconv.Atoi(m)
			if err != nil {
				return zeroDur, err
			}
			d.Months += val
		case "week":
			val, err := strconv.Atoi(m)
			if err != nil {
				return zeroDur, err
			}
			d.Days += (val * 7)
		case "day":
			val, err := strconv.Atoi(m)
			if err != nil {
				return zeroDur, err
			}
			d.Days += val
		case "hour":
			val, err := time.ParseDuration(m + "h")
			if err != nil {
				return zeroDur, err
			}
			d.TimePart += val
		case "minute":
			val, err := time.ParseDuration(m + "m")
			if err != nil {
				return zeroDur, err
			}
			d.TimePart += val
		case "second":
			val, err := time.ParseDuration(strings.ReplaceAll(m, ",", ".") + "s")
			if err != nil {
				return zeroDur, err
			}
			d.TimePart += val
		default:
			return zeroDur, fmt.Errorf("unknown field %s", name)
		}
	}

	return d, err
}
