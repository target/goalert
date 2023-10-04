package timeutil

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/validation"
)

// ISORInterval represents an ISO recurring interval.
type ISORInterval struct {
	Repeat int
	Start  time.Time
	Period ISODuration
}

func (r *ISORInterval) calcStart(end time.Time) error {
	if r.Period.IsZero() {
		return fmt.Errorf("invalid interval: duration must be non-zero")
	}
	n := r.Repeat + 1
	r.Start = end.AddDate(-r.Period.YearPart*n, -r.Period.MonthPart*n, -r.Period.Days()*n).Add(-r.Period.TimePart() * time.Duration(n))
	return nil
}

func (r *ISORInterval) calcPeriod(end time.Time) error {
	if !end.After(r.Start) {
		return fmt.Errorf("invalid interval: end time must be after start time: %s", end.Format(time.RFC3339Nano))
	}

	r.Period.SetTimePart(end.Sub(r.Start) / time.Duration(r.Repeat+1))

	return nil
}

// End returns the end time of the interval.
func (r ISORInterval) End() time.Time {
	if r.Repeat < 0 {
		panic("cannot calculate end time for infinite interval")
	}

	n := r.Repeat + 1
	return r.Start.UTC().AddDate(r.Period.YearPart*n, r.Period.MonthPart*n, r.Period.Days()*n).Add(r.Period.TimePart() * time.Duration(n))
}

// String returns the string representation of the interval.
func (r ISORInterval) String() string {
	if r.Period.YearPart == 0 && r.Period.MonthPart == 0 && r.Period.Days() == 0 {
		// just a duration, use start/end format
		return fmt.Sprintf("R%d/%s/%s", r.Repeat, r.Start.Format(time.RFC3339Nano), r.End().Format(time.RFC3339Nano))
	}

	return fmt.Sprintf("R%d/%s/%s", r.Repeat, r.Start.Format(time.RFC3339Nano), r.Period.String())
}

// ParseISORInterval parses an ISO recurring interval string. If the string has a duration only,
// the start time will be set to the current time. Infinite intervals are not supported.
func ParseISORInterval(s string) (ISORInterval, error) {
	return ParseISORIntervalFrom(time.Now(), s)
}

// ParseISORIntervalFrom parses an ISO recurring interval string. If the string has a duration only,
// the start time will be set to t. Infinite intervals are not supported.
func ParseISORIntervalFrom(t time.Time, s string) (ISORInterval, error) {
	parts := strings.SplitN(s, "/", 3)
	if len(parts) < 2 {
		return ISORInterval{}, fmt.Errorf("invalid interval: %s", s)
	}

	if parts[0][0] != 'R' {
		return ISORInterval{}, fmt.Errorf("invalid interval: missing R value: %s", s)
	}

	var ivl ISORInterval

	var err error
	ivl.Repeat, err = strconv.Atoi(parts[0][1:])
	if err != nil {
		return ISORInterval{}, fmt.Errorf("invalid interval: invalid R value: %s", parts[0][1:])
	}
	if ivl.Repeat < 0 {
		return ISORInterval{}, fmt.Errorf("invalid interval: R value must be positive: %d", ivl.Repeat)
	}

	var hasStart bool
	if parts[1][0] == 'P' {
		ivl.Period, err = ParseISODuration(parts[1])
		if err != nil {
			return ISORInterval{}, fmt.Errorf("invalid interval: invalid duration: %s", parts[1])
		}
	} else {
		hasStart = true
		ivl.Start, err = time.Parse(time.RFC3339Nano, parts[1])
		if err != nil {
			return ISORInterval{}, fmt.Errorf("invalid interval: invalid start time: %s", parts[1])
		}
	}

	if len(parts) == 2 {
		// just a duration and count, use provided start time
		if !hasStart {
			ivl.Start = t
		}
		return ivl, nil
	}

	if parts[2][0] == 'P' && !hasStart {
		return ISORInterval{}, fmt.Errorf("invalid interval: got two durations, expected end time: %s", s)
	}

	if parts[2][0] == 'P' {
		ivl.Period, err = ParseISODuration(parts[2])
		if err != nil {
			return ISORInterval{}, fmt.Errorf("invalid interval: invalid duration: %s", parts[2])
		}
		return ivl, nil
	}

	end, err := time.Parse(time.RFC3339Nano, parts[2])
	if err != nil {
		return ISORInterval{}, fmt.Errorf("invalid interval: invalid end time: %s", parts[2])
	}
	if hasStart {
		err = ivl.calcPeriod(end)
	} else {
		err = ivl.calcStart(end)
	}

	return ivl, err
}

func (r ISORInterval) MarshalGQL(w io.Writer) {
	if r == (ISORInterval{}) {
		_, _ = io.WriteString(w, "null")
		return
	}

	_, _ = io.WriteString(w, `"`+r.String()+`"`)
}

func (r *ISORInterval) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return errors.New("ISORIntervals must be strings")
	}
	str = strings.Trim(str, `"`)

	t, err := ParseISORInterval(str)
	if err != nil {
		return validation.WrapError(err)
	}

	*r = t
	return nil
}
