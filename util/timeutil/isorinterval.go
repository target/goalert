package timeutil

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ISORInterval represents an ISO recurring interval.
type ISORInterval struct {
	Count  int
	Start  time.Time
	Period ISODuration
}

func (r *ISORInterval) calcStart(end time.Time) {
	n := r.Count + 1
	r.Start = end.AddDate(-r.Period.Years*n, -r.Period.Months*n, -r.Period.Days*n).Add(-r.Period.TimePart * time.Duration(n))
}

func (r *ISORInterval) calcPeriod(end time.Time) error {
	if !end.After(r.Start) {
		return fmt.Errorf("invalid interval: end time must be after start time: %s", end.Format(time.RFC3339Nano))
	}

	r.Period.TimePart = end.Sub(r.Start) / time.Duration(r.Count+1)

	return nil
}

// End returns the end time of the interval.
func (r ISORInterval) End() time.Time {
	if r.Count < 0 {
		panic("cannot calculate end time for infinite interval")
	}

	n := r.Count + 1
	return r.Start.UTC().AddDate(r.Period.Years*n, r.Period.Months*n, r.Period.Days*n).Add(r.Period.TimePart * time.Duration(n))
}

// String returns the string representation of the interval.
func (r ISORInterval) String() string {
	if r.Period.Years == 0 && r.Period.Months == 0 && r.Period.Days == 0 {
		// just a duration, use start/end format
		return fmt.Sprintf("R%d/%s/%s", r.Count, r.Start.Format(time.RFC3339Nano), r.End().Format(time.RFC3339Nano))
	}

	return fmt.Sprintf("R%d/%s/%s", r.Count, r.Start.Format(time.RFC3339Nano), r.Period.String())
}

// ParseRInterval parses an ISO recurring interval string. If the string has a duration only,
// the start time will be set to the current time. Infinite intervals are not supported.
func ParseISORInterval(s string) (ISORInterval, error) {
	return ParseISORIntervalFrom(time.Now(), s)
}

// ParseRInterval parses an ISO recurring interval string. If the string has a duration only,
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
	ivl.Count, err = strconv.Atoi(parts[0][1:])
	if err != nil {
		return ISORInterval{}, fmt.Errorf("invalid interval: invalid R value: %s", parts[0][1:])
	}
	if ivl.Count < 0 {
		return ISORInterval{}, fmt.Errorf("invalid interval: R value must be positive: %d", ivl.Count)
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
		ivl.calcStart(end)
	}

	return ivl, err
}
