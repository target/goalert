package timeutil

import (
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"github.com/target/goalert/validation"
)

// ISODuration represents an ISO duration string.
// It is a subset of the ISO 8601 Durations format (https://en.wikipedia.org/wiki/ISO_8601#Durations).
//
// Notably, it does not support negative values, and fractional values are only supported for seconds.
type ISODuration struct {
	YearPart   int
	MonthPart  int
	WeekPart   int
	DayPart    int
	HourPart   int
	MinutePart int
	SecondPart float64
}

// ISODurationFromTime returns an ISODuration with the given time.Duration as the time part.
func ISODurationFromTime(t time.Duration) ISODuration {
	var dur ISODuration
	dur.SetTimePart(t)
	return dur
}

// PGXInterval returns a pgtype.Interval representation of the duration.
func (dur ISODuration) PGXInterval() pgtype.Interval {
	return pgtype.Interval{
		Microseconds: int64(dur.TimePart() / time.Microsecond),
		Days:         int32(dur.Days()),
		Months:       int32(dur.MonthPart + dur.YearPart*12),
		Valid:        true,
	}
}

// Days returns the total number of days in the duration (adds DayPart and WeekPart appropriately).
func (dur ISODuration) Days() int {
	return dur.DayPart + (dur.WeekPart * 7)
}

// TimePart returns the time portion of the duration as a time.Duration.
func (dur ISODuration) TimePart() time.Duration {
	return time.Duration(dur.HourPart)*time.Hour + time.Duration(dur.MinutePart)*time.Minute + time.Duration(dur.SecondPart*float64(time.Second))
}

// SetTimePart sets the time portion of the duration from a time.Duration.
func (dur *ISODuration) SetTimePart(timeDur time.Duration) {
	dur.HourPart = int(timeDur.Hours())
	dur.MinutePart = int(timeDur.Minutes()) % 60
	dur.SecondPart = timeDur.Seconds() - float64(dur.HourPart*60*60+dur.MinutePart*60)
}

var zeroDur ISODuration

func (dur ISODuration) IsZero() bool {
	return dur == zeroDur
}

// AddTo adds the duration to the given time.
func (dur ISODuration) AddTo(t time.Time) time.Time {
	return t.AddDate(dur.YearPart, dur.MonthPart, dur.Days()).Add(dur.TimePart())
}

// LessThan returns true if the duration is less than the other duration from the given reference time.
func (dur ISODuration) LessThan(t time.Time, other ISODuration) bool {
	return dur.AddTo(t).Before(other.AddTo(t))
}

// Equal returns true if the duration is equal to the other duration from the given reference time.
func (dur ISODuration) Equal(t time.Time, other ISODuration) bool {
	return dur.AddTo(t).Equal(other.AddTo(t))
}

// String returns an ISO 8601 duration string, rounded to the nearest microsecond.
func (dur ISODuration) String() string {
	if dur == zeroDur {
		return "P0D"
	}

	var b strings.Builder
	b.WriteRune('P')

	if dur.YearPart > 0 {
		fmt.Fprintf(&b, "%dY", dur.YearPart)
	}
	if dur.MonthPart > 0 {
		fmt.Fprintf(&b, "%dM", dur.MonthPart)
	}
	if dur.WeekPart > 0 {
		fmt.Fprintf(&b, "%dW", dur.WeekPart)
	}
	if dur.DayPart > 0 {
		fmt.Fprintf(&b, "%dD", dur.DayPart)
	}

	if dur.SecondPart <= 0 && dur.MinutePart <= 0 && dur.HourPart <= 0 {
		return b.String()
	}

	b.WriteRune('T')

	if dur.HourPart > 0 {
		fmt.Fprintf(&b, "%dH", dur.HourPart)
	}

	if dur.MinutePart > 0 {
		fmt.Fprintf(&b, "%dM", dur.MinutePart)
	}

	if dur.SecondPart > 0 {
		sec := dur.SecondPart
		// round to microseconds
		sec = math.Round(sec*1e6) / 1e6
		fmt.Fprintf(&b, "%sS", strconv.FormatFloat(sec, 'f', -1, 64))
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
			d.YearPart = val
		case "month":
			val, err := strconv.Atoi(m)
			if err != nil {
				return zeroDur, err
			}
			d.MonthPart = val
		case "week":
			val, err := strconv.Atoi(m)
			if err != nil {
				return zeroDur, err
			}
			d.WeekPart = val
		case "day":
			val, err := strconv.Atoi(m)
			if err != nil {
				return zeroDur, err
			}
			d.DayPart = val
		case "hour":
			val, err := strconv.Atoi(m)
			if err != nil {
				return zeroDur, err
			}
			d.HourPart = val
		case "minute":
			val, err := strconv.Atoi(m)
			if err != nil {
				return zeroDur, err
			}
			d.MinutePart = val
		case "second":
			val, err := strconv.ParseFloat(strings.ReplaceAll(m, ",", "."), 64)
			if err != nil {
				return zeroDur, err
			}
			d.SecondPart = val
		default:
			return zeroDur, fmt.Errorf("unknown field %s", name)
		}
	}

	return d, err
}

func (dur ISODuration) MarshalGQL(w io.Writer) {
	if dur == (ISODuration{}) {
		_, _ = io.WriteString(w, "null")
		return
	}

	_, _ = io.WriteString(w, `"`+dur.String()+`"`)
}

func (dur *ISODuration) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return errors.New("ISORIntervals must be strings")
	}
	str = strings.Trim(str, `"`)

	t, err := ParseISODuration(str)
	if err != nil {
		return validation.WrapError(err)
	}

	*dur = t
	return nil
}
