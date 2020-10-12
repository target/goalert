package schedule

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTemporarySchedule_TrimEnd(t *testing.T) {

	temp := TemporarySchedule{
		Start: time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC),
		End:   time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC),
		Shifts: []FixedShift{
			{
				Start:  time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
				End:    time.Date(2000, 1, 1, 18, 0, 0, 0, time.UTC),
				UserID: "foo",
			},
			{
				Start:  time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
				End:    time.Date(2000, 1, 1, 15, 0, 0, 0, time.UTC),
				UserID: "bar",
			},
		},
	}

	res := temp.TrimEnd(time.Date(2000, 1, 1, 14, 0, 0, 0, time.UTC))

	assert.EqualValues(t, TemporarySchedule{

		Start: time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC),
		End:   time.Date(2000, 1, 1, 14, 0, 0, 0, time.UTC),
		Shifts: []FixedShift{
			{
				Start:  time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
				End:    time.Date(2000, 1, 1, 14, 0, 0, 0, time.UTC),
				UserID: "foo",
			},
			{
				Start:  time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
				End:    time.Date(2000, 1, 1, 14, 0, 0, 0, time.UTC),
				UserID: "bar",
			},
		},
	}, res)
}

func TestTemporarySchedule_TrimStart(t *testing.T) {

	temp := TemporarySchedule{
		Start: time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC),
		End:   time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC),
		Shifts: []FixedShift{
			{
				Start:  time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
				End:    time.Date(2000, 1, 1, 18, 0, 0, 0, time.UTC),
				UserID: "foo",
			},
			{
				Start:  time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
				End:    time.Date(2000, 1, 1, 15, 0, 0, 0, time.UTC),
				UserID: "bar",
			},
		},
	}

	res := temp.TrimStart(time.Date(2000, 1, 1, 16, 0, 0, 0, time.UTC))

	assert.EqualValues(t, TemporarySchedule{
		Start: time.Date(2000, 1, 1, 16, 0, 0, 0, time.UTC),
		End:   time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC),
		Shifts: []FixedShift{
			{
				Start:  time.Date(2000, 1, 1, 16, 0, 0, 0, time.UTC),
				End:    time.Date(2000, 1, 1, 18, 0, 0, 0, time.UTC),
				UserID: "foo",
			},
		},
	}, res)
}

func TestMergeTemporarySchedules(t *testing.T) {

	tempScheds := []TemporarySchedule{{
		Start: time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC),
		End:   time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC),
		Shifts: []FixedShift{
			{
				Start:  time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
				End:    time.Date(2000, 1, 1, 18, 0, 0, 0, time.UTC),
				UserID: "foo",
			},
			{
				Start:  time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
				End:    time.Date(2000, 1, 1, 15, 0, 0, 0, time.UTC),
				UserID: "bar",
			},
			{
				Start:  time.Date(2000, 1, 1, 11, 0, 0, 0, time.UTC),
				End:    time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
				UserID: "foo",
			},
			{
				Start:  time.Date(2000, 1, 1, 14, 0, 0, 0, time.UTC),
				End:    time.Date(2000, 1, 1, 16, 0, 0, 0, time.UTC),
				UserID: "bar",
			},
		},
	}, {
		Start: time.Date(2000, 1, 1, 9, 0, 0, 0, time.UTC),
		End:   time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC),
		Shifts: []FixedShift{
			{
				Start:  time.Date(2000, 1, 1, 11, 0, 0, 0, time.UTC),
				End:    time.Date(2000, 1, 1, 15, 0, 0, 0, time.UTC),
				UserID: "bar",
			},
		},
	}, {
		Start: time.Date(2000, 1, 1, 21, 0, 0, 0, time.UTC),
		End:   time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC),
		Shifts: []FixedShift{
			{
				Start:  time.Date(2000, 1, 1, 21, 0, 0, 0, time.UTC),
				End:    time.Date(2000, 1, 1, 23, 0, 0, 0, time.UTC),
				UserID: "bar",
			},
		},
	},
	}

	res := MergeTemporarySchedules(tempScheds)

	assert.EqualValues(t, []TemporarySchedule{
		{
			Start: time.Date(2000, 1, 1, 9, 0, 0, 0, time.UTC),
			End:   time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC),
			Shifts: []FixedShift{
				{
					Start:  time.Date(2000, 1, 1, 11, 0, 0, 0, time.UTC),
					End:    time.Date(2000, 1, 1, 16, 0, 0, 0, time.UTC),
					UserID: "bar",
				},
				{
					Start:  time.Date(2000, 1, 1, 11, 0, 0, 0, time.UTC),
					End:    time.Date(2000, 1, 1, 18, 0, 0, 0, time.UTC),
					UserID: "foo",
				},
			},
		},
		{
			Start: time.Date(2000, 1, 1, 21, 0, 0, 0, time.UTC),
			End:   time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC),
			Shifts: []FixedShift{
				{
					Start:  time.Date(2000, 1, 1, 21, 0, 0, 0, time.UTC),
					End:    time.Date(2000, 1, 1, 22, 0, 0, 0, time.UTC),
					UserID: "bar",
				},
			},
		},
	}, res)
}

func TestDeleteFixedShifts(t *testing.T) {

	tempScheds := []TemporarySchedule{
		{
			Start: time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC),
			End:   time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC),
			Shifts: []FixedShift{
				{
					Start:  time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
					End:    time.Date(2000, 1, 1, 18, 0, 0, 0, time.UTC),
					UserID: "foo",
				},
				{
					Start:  time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
					End:    time.Date(2000, 1, 1, 15, 0, 0, 0, time.UTC),
					UserID: "bar",
				},
			},
		},
	}

	res := deleteFixedShifts(tempScheds, time.Date(2000, 1, 1, 14, 0, 0, 0, time.UTC), time.Date(2000, 1, 1, 16, 0, 0, 0, time.UTC))

	assert.EqualValues(t, []TemporarySchedule{
		{
			Start: time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC),
			End:   time.Date(2000, 1, 1, 14, 0, 0, 0, time.UTC),
			Shifts: []FixedShift{
				{
					Start:  time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
					End:    time.Date(2000, 1, 1, 14, 0, 0, 0, time.UTC),
					UserID: "foo",
				},
				{
					Start:  time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
					End:    time.Date(2000, 1, 1, 14, 0, 0, 0, time.UTC),
					UserID: "bar",
				},
			},
		},
		{
			Start: time.Date(2000, 1, 1, 16, 0, 0, 0, time.UTC),
			End:   time.Date(2000, 1, 1, 20, 0, 0, 0, time.UTC),
			Shifts: []FixedShift{
				{
					Start:  time.Date(2000, 1, 1, 16, 0, 0, 0, time.UTC),
					End:    time.Date(2000, 1, 1, 18, 0, 0, 0, time.UTC),
					UserID: "foo",
				},
			},
		},
	}, res)
}
