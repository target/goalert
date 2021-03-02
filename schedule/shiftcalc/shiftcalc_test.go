package shiftcalc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/util/timeutil"
)

func parseDate(t *testing.T, value string) time.Time {
	t.Helper()
	return parseTimeFmt(t, "Jan _2 3:04PM 2006", value)
}

func parseTimeFmt(t *testing.T, layout, value string) time.Time {
	t.Helper()
	tm, err := time.ParseInLocation(layout, value, time.Local)
	if err != nil {
		t.Fatal(err)
	}
	return tm.In(time.Local)
}

func TestRotationShifts(t *testing.T) {
	rotStart := time.Date(2017, time.January, 0, 1, 0, 0, 0, time.UTC)
	qStart := time.Date(2017, time.January, 0, 0, 0, 0, 0, time.UTC)
	qEnd := time.Date(2017, time.February, 0, 0, 0, 0, 0, time.UTC)

	rot := &rotation.Rotation{
		Type:        rotation.TypeDaily,
		ShiftLength: 1,
		Start:       rotStart,
	}
	parts := []string{"first", "second", "third"}
	shifts := _rotationShifts(qStart, qEnd, rot, 1, rotStart, parts)

	if len(shifts) != 31 {
		t.Errorf("got %d shifts; want 31", len(shifts))
	}
	if shifts[0].PartID != "second" {
		t.Errorf("got '%s' participant for first shift; want 'second' participant", shifts[0].PartID)
	}
	if !shifts[0].Start.Equal(rotStart) {
		t.Errorf("got '%s' for first shift start; want rotation start (%s)", shifts[0].Start.String(), rotStart.String())
	}
}

func TestRuleShifts(t *testing.T) {
	start := parseDate(t, "Jul 20 11:00AM 2017")
	end := parseDate(t, "Jul 24 11:00AM 2017")

	var r rule.Rule
	r.Start = timeutil.NewClock(8, 0)
	r.End = timeutil.NewClock(20, 0)
	r.SetDay(time.Friday, true)
	r.SetDay(time.Saturday, true)
	r.SetDay(time.Monday, true)

	shifts := ruleShifts(start, end, r)

	assert.Contains(t, shifts, Shift{
		Start: parseDate(t, "Jul 21 8:00AM 2017"),
		End:   parseDate(t, "Jul 21 8:00PM 2017"),
	})
	assert.Contains(t, shifts, Shift{
		Start: parseDate(t, "Jul 22 8:00AM 2017"),
		End:   parseDate(t, "Jul 22 8:00PM 2017"),
	})
	assert.Contains(t, shifts, Shift{
		Start: parseDate(t, "Jul 24 8:00AM 2017"),
		End:   parseDate(t, "Jul 24 8:00PM 2017"),
	})

	assert.Len(t, shifts, 3)
}

// for historical shift data
func TestHistoricalShifts(t *testing.T) {
	start := parseDate(t, "Jul 20 11:00AM 2018")
	end := parseDate(t, "Aug 24 11:00AM 2018")

	var r rule.Rule
	r.Start = timeutil.NewClock(8, 0)
	r.End = timeutil.NewClock(20, 0)
	r.SetDay(time.Sunday, true)
	r.SetDay(time.Monday, true)
	r.SetDay(time.Tuesday, true)

	shifts := ruleShifts(start, end, r)

	assert.Contains(t, shifts, Shift{
		Start: parseDate(t, "Jul 23 8:00AM 2018"),
		End:   parseDate(t, "Jul 23 8:00PM 2018"),
	})
	assert.Contains(t, shifts, Shift{
		Start: parseDate(t, "Aug 20 8:00AM 2018"),
		End:   parseDate(t, "Aug 20 8:00PM 2018"),
	})
	assert.NotContains(t, shifts, Shift{
		Start: parseDate(t, "Aug 22 11:00AM 2018"),
		End:   parseDate(t, "Aug 22 11:00PM 2018"),
	})
	assert.NotContains(t, shifts, Shift{
		Start: parseDate(t, "Jul 19 11:00AM 2018"),
		End:   parseDate(t, "Jul 19 11:00PM 2018"),
	})
	assert.NotContains(t, shifts, Shift{
		Start: parseDate(t, "Jul 19 11:00AM 2018"),
		End:   parseDate(t, "Jul 20 11:00PM 2018"),
	})
	assert.NotContains(t, shifts, Shift{
		Start: parseDate(t, "Aug 24 11:00AM 2018"),
		End:   parseDate(t, "Aug 29 11:00PM 2018"),
	})

	assert.Len(t, shifts, 15)

}
