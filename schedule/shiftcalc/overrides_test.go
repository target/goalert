package shiftcalc

import (
	"github.com/target/goalert/override"
	"testing"
	"time"
)

func TestFinalShiftsWithOverrides(t *testing.T) {
	check := func(shifts []Shift, overrides []override.UserOverride, expected []Shift) {
		t.Run("", func(t *testing.T) {
			res := finalShiftsWithOverrides(shifts, overrides)
			if len(res) != len(expected) {
				t.Fatalf("got len() = %d; want %d", len(res), len(expected))
			}
			for i, exp := range expected {
				if !res[i].Start.Equal(exp.Start) {
					t.Errorf("Shift[%d]: Start=%s; want %s", i, res[i].Start, exp.Start)
				}
				if !res[i].End.Equal(exp.End) {
					t.Errorf("Shift[%d]: End=%s; want %s", i, res[i].End, exp.End)
				}
				if res[i].UserID != exp.UserID {
					t.Errorf("Shift[%d]: UserID=%s; want %s", i, res[i].UserID, exp.UserID)
				}
			}
		})
	}

	check(
		[]Shift{
			{UserID: "a", Start: time.Date(0, 0, 0, 0, 9, 0, 0, time.UTC), End: time.Date(0, 0, 0, 0, 21, 0, 0, time.UTC)},
			{UserID: "a", Start: time.Date(0, 0, 1, 0, 9, 0, 0, time.UTC), End: time.Date(0, 0, 1, 0, 21, 0, 0, time.UTC)},
		},
		nil,
		[]Shift{
			{UserID: "a", Start: time.Date(0, 0, 0, 0, 9, 0, 0, time.UTC), End: time.Date(0, 0, 0, 0, 21, 0, 0, time.UTC)},
			{UserID: "a", Start: time.Date(0, 0, 1, 0, 9, 0, 0, time.UTC), End: time.Date(0, 0, 1, 0, 21, 0, 0, time.UTC)},
		},
	)

	check(
		[]Shift{
			{UserID: "a", Start: time.Date(0, 0, 0, 0, 9, 0, 0, time.UTC), End: time.Date(0, 0, 0, 0, 21, 0, 0, time.UTC)},
			{UserID: "a", Start: time.Date(0, 0, 1, 0, 9, 0, 0, time.UTC), End: time.Date(0, 0, 1, 0, 21, 0, 0, time.UTC)},
		},
		[]override.UserOverride{{RemoveUserID: "a", Start: time.Date(0, 0, 0, 0, 9, 0, 0, time.UTC), End: time.Date(0, 0, 0, 0, 21, 0, 0, time.UTC)}},
		[]Shift{
			{UserID: "a", Start: time.Date(0, 0, 1, 0, 9, 0, 0, time.UTC), End: time.Date(0, 0, 1, 0, 21, 0, 0, time.UTC)},
		},
	)

	check(
		[]Shift{
			{UserID: "a", Start: time.Date(0, 0, 0, 0, 9, 0, 0, time.UTC), End: time.Date(0, 0, 0, 0, 21, 0, 0, time.UTC)},
			{UserID: "a", Start: time.Date(0, 0, 1, 0, 9, 0, 0, time.UTC), End: time.Date(0, 0, 1, 0, 21, 0, 0, time.UTC)},
		},
		[]override.UserOverride{{AddUserID: "b", Start: time.Date(0, 0, 0, 0, 9, 0, 0, time.UTC), End: time.Date(0, 0, 0, 0, 21, 0, 0, time.UTC)}},
		[]Shift{
			{UserID: "a", Start: time.Date(0, 0, 0, 0, 9, 0, 0, time.UTC), End: time.Date(0, 0, 0, 0, 21, 0, 0, time.UTC)},
			{UserID: "b", Start: time.Date(0, 0, 0, 0, 9, 0, 0, time.UTC), End: time.Date(0, 0, 0, 0, 21, 0, 0, time.UTC)},
			{UserID: "a", Start: time.Date(0, 0, 1, 0, 9, 0, 0, time.UTC), End: time.Date(0, 0, 1, 0, 21, 0, 0, time.UTC)},
		},
	)

	check(
		[]Shift{
			{UserID: "a", Start: time.Date(0, 0, 0, 0, 9, 0, 0, time.UTC), End: time.Date(0, 0, 0, 0, 21, 0, 0, time.UTC)},
			{UserID: "a", Start: time.Date(0, 0, 1, 0, 9, 0, 0, time.UTC), End: time.Date(0, 0, 1, 0, 21, 0, 0, time.UTC)},
		},
		[]override.UserOverride{{RemoveUserID: "a", AddUserID: "b", Start: time.Date(0, 0, 0, 0, 9, 0, 0, time.UTC), End: time.Date(0, 0, 0, 0, 9, 5, 0, time.UTC)}},
		[]Shift{
			{UserID: "b", Start: time.Date(0, 0, 0, 0, 9, 0, 0, time.UTC), End: time.Date(0, 0, 0, 0, 9, 5, 0, time.UTC)},
			{UserID: "a", Start: time.Date(0, 0, 0, 0, 9, 5, 0, time.UTC), End: time.Date(0, 0, 0, 0, 21, 0, 0, time.UTC)},
			{UserID: "a", Start: time.Date(0, 0, 1, 0, 9, 0, 0, time.UTC), End: time.Date(0, 0, 1, 0, 21, 0, 0, time.UTC)},
		},
	)

	check(
		[]Shift{
			{UserID: "Joey", Start: time.Date(2018, 3, 16, 13, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 17, 1, 30, 0, 0, time.UTC)},
			{UserID: "Srilekha", Start: time.Date(2018, 3, 17, 1, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 17, 13, 30, 0, 0, time.UTC)},
			{UserID: "Joey", Start: time.Date(2018, 3, 17, 13, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 18, 1, 30, 0, 0, time.UTC)},
			{UserID: "Srilekha", Start: time.Date(2018, 3, 18, 1, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 18, 13, 30, 0, 0, time.UTC)},
			{UserID: "Joey", Start: time.Date(2018, 3, 18, 13, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 19, 1, 30, 0, 0, time.UTC)},
			{UserID: "Srilekha", Start: time.Date(2018, 3, 19, 1, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 19, 13, 30, 0, 0, time.UTC)},
			{UserID: "Joey", Start: time.Date(2018, 3, 19, 13, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 20, 1, 30, 0, 0, time.UTC)},
			{UserID: "Srilekha", Start: time.Date(2018, 3, 20, 1, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 20, 13, 30, 0, 0, time.UTC)},
			{UserID: "Joey", Start: time.Date(2018, 3, 20, 13, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 21, 1, 30, 0, 0, time.UTC)},
			{UserID: "Srilekha", Start: time.Date(2018, 3, 21, 1, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 21, 13, 30, 0, 0, time.UTC)},
			{UserID: "Joey", Start: time.Date(2018, 3, 21, 13, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 22, 1, 30, 0, 0, time.UTC)},
			{UserID: "Srilekha", Start: time.Date(2018, 3, 22, 1, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 22, 13, 30, 0, 0, time.UTC)},
		},
		[]override.UserOverride{
			{AddUserID: "Tom", RemoveUserID: "Joey", Start: time.Date(2018, 3, 17, 22, 0, 0, 0, time.UTC), End: time.Date(2018, 3, 18, 1, 30, 0, 0, time.UTC)},
			{AddUserID: "Dyanne", RemoveUserID: "Joey", Start: time.Date(2018, 3, 20, 13, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 21, 1, 30, 0, 0, time.UTC)},
			{AddUserID: "Tom", RemoveUserID: "Joey", Start: time.Date(2018, 3, 19, 19, 0, 0, 0, time.UTC), End: time.Date(2018, 3, 20, 1, 30, 0, 0, time.UTC)},
		},
		[]Shift{
			{UserID: "Joey", Start: time.Date(2018, 3, 16, 13, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 17, 1, 30, 0, 0, time.UTC)},
			{UserID: "Srilekha", Start: time.Date(2018, 3, 17, 1, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 17, 13, 30, 0, 0, time.UTC)},

			// Tom takes over the end of Joey's shift
			{UserID: "Joey", Start: time.Date(2018, 3, 17, 13, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 17, 22, 00, 0, 0, time.UTC)},
			{UserID: "Tom", Start: time.Date(2018, 3, 17, 22, 00, 0, 0, time.UTC), End: time.Date(2018, 3, 18, 1, 30, 0, 0, time.UTC)},

			{UserID: "Srilekha", Start: time.Date(2018, 3, 18, 1, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 18, 13, 30, 0, 0, time.UTC)},
			{UserID: "Joey", Start: time.Date(2018, 3, 18, 13, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 19, 1, 30, 0, 0, time.UTC)},
			{UserID: "Srilekha", Start: time.Date(2018, 3, 19, 1, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 19, 13, 30, 0, 0, time.UTC)},

			// Tom takes over the end of Joey's shift
			{UserID: "Joey", Start: time.Date(2018, 3, 19, 13, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 19, 19, 0, 0, 0, time.UTC)},
			{UserID: "Tom", Start: time.Date(2018, 3, 19, 19, 0, 0, 0, time.UTC), End: time.Date(2018, 3, 20, 1, 30, 0, 0, time.UTC)},

			{UserID: "Srilekha", Start: time.Date(2018, 3, 20, 1, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 20, 13, 30, 0, 0, time.UTC)},

			// Dyanne takes over Joey's entire shift
			{UserID: "Dyanne", Start: time.Date(2018, 3, 20, 13, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 21, 1, 30, 0, 0, time.UTC)},

			{UserID: "Srilekha", Start: time.Date(2018, 3, 21, 1, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 21, 13, 30, 0, 0, time.UTC)},
			{UserID: "Joey", Start: time.Date(2018, 3, 21, 13, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 22, 1, 30, 0, 0, time.UTC)},
			{UserID: "Srilekha", Start: time.Date(2018, 3, 22, 1, 30, 0, 0, time.UTC), End: time.Date(2018, 3, 22, 13, 30, 0, 0, time.UTC)},
		},
	)

}
