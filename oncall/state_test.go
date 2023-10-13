package oncall

import (
	"testing"
	"time"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/override"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/util/timeutil"
)

func BenchmarkState_CalculateShifts(b *testing.B) {
	s := &state{
		loc: time.UTC,
		rules: []ResolvedRule{
			{Rule: rule.Rule{
				WeekdayFilter: timeutil.WeekdayFilter{1, 1, 1, 1, 1, 1, 1},
				Start:         timeutil.NewClock(8, 0),
				End:           timeutil.NewClock(10, 0),
				Target:        assignment.UserTarget("foobar"),
			}},
			{Rule: rule.Rule{
				WeekdayFilter: timeutil.WeekdayFilter{1, 1, 1, 1, 1, 1, 1},
				Start:         timeutil.NewClock(8, 0),
				End:           timeutil.NewClock(10, 0),
				Target:        assignment.UserTarget("foobar"),
			}},
			{Rule: rule.Rule{
				WeekdayFilter: timeutil.WeekdayFilter{1, 0, 1, 1, 1, 1, 1},
				Start:         timeutil.NewClock(8, 0),
				End:           timeutil.NewClock(10, 0),
				Target:        assignment.UserTarget("foobar"),
			}},
			{Rule: rule.Rule{
				WeekdayFilter: timeutil.WeekdayFilter{1, 1, 1, 1, 0, 1, 1},
				Start:         timeutil.NewClock(8, 0),
				End:           timeutil.NewClock(10, 0),
				Target:        assignment.UserTarget("foobar2"),
			}},
			{Rule: rule.Rule{
				WeekdayFilter: timeutil.WeekdayFilter{1, 1, 1, 1, 1, 1, 1},
				Start:         timeutil.NewClock(8, 0),
				End:           timeutil.NewClock(10, 0),
				Target:        assignment.UserTarget("fooba4r"),
			}},
			{
				Rule: rule.Rule{
					WeekdayFilter: timeutil.WeekdayFilter{1, 1, 1, 1, 1, 1, 1},
					Start:         timeutil.NewClock(8, 0),
					End:           timeutil.NewClock(10, 0),
					Target:        assignment.RotationTarget("fooba4r"),
				},
				Rotation: &ResolvedRotation{
					Rotation: rotation.Rotation{
						Type:        rotation.TypeDaily,
						ShiftLength: 2,
						Start:       time.Date(2017, 1, 2, 3, 4, 5, 6, time.UTC),
					},
					Users: []string{"a", "b", "c", "d", "e"},
				},
			},
			{
				Rule: rule.Rule{
					WeekdayFilter: timeutil.WeekdayFilter{1, 1, 1, 1, 1, 1, 1},
					Start:         timeutil.NewClock(8, 0),
					End:           timeutil.NewClock(10, 0),
					Target:        assignment.RotationTarget("fooba4r"),
				},
				Rotation: &ResolvedRotation{
					Rotation: rotation.Rotation{
						Type:        rotation.TypeDaily,
						ShiftLength: 2,
						Start:       time.Date(2017, 1, 2, 3, 4, 5, 6, time.UTC),
					},
					Users: []string{"a", "b", "c", "d", "e"},
				},
			},
			{
				Rule: rule.Rule{
					WeekdayFilter: timeutil.WeekdayFilter{1, 1, 1, 1, 1, 1, 1},
					Start:         timeutil.NewClock(8, 0),
					End:           timeutil.NewClock(10, 0),
					Target:        assignment.RotationTarget("fooba4r"),
				},
				Rotation: &ResolvedRotation{
					Rotation: rotation.Rotation{
						Type:        rotation.TypeDaily,
						ShiftLength: 2,
						Start:       time.Date(2017, 1, 2, 3, 4, 5, 6, time.UTC),
					},
					Users: []string{"a", "b", "c", "d", "e"},
				},
			},
		},
		overrides: []override.UserOverride{
			{
				AddUserID:    "binbaz",
				RemoveUserID: "foobar",
				Start:        time.Date(2018, 1, 1, 8, 30, 0, 0, time.UTC),
				End:          time.Date(2018, 1, 1, 8, 45, 0, 0, time.UTC),
			},
			{
				AddUserID:    "binbaz2",
				RemoveUserID: "foobar",
				Start:        time.Date(2018, 1, 1, 8, 30, 0, 0, time.UTC),
				End:          time.Date(2018, 1, 1, 8, 45, 0, 0, time.UTC),
			},
			{
				AddUserID:    "binbaz",
				RemoveUserID: "foob3ar",
				Start:        time.Date(2018, 1, 1, 8, 30, 0, 0, time.UTC),
				End:          time.Date(2018, 1, 1, 8, 45, 0, 0, time.UTC),
			},
		},
	}
	s.CalculateShifts(
		time.Date(2018, 1, 1, 8, 0, 0, 0, time.UTC), // 8:00AM
		time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC), // 9:00AM
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.CalculateShifts(
			time.Date(2018, 1, 1, 8, 0, 0, 0, time.UTC), // 8:00AM
			time.Date(2019, 1, 1, 8, 0, 0, 0, time.UTC), // 9:00AM
		)
	}
}

func TestResolvedRotation_UserID(t *testing.T) {
	rot := &ResolvedRotation{
		Rotation: rotation.Rotation{
			ID:          "rot",
			Type:        rotation.TypeWeekly,
			Start:       time.Date(2018, 4, 21, 14, 8, 9, 168379000, time.UTC),
			ShiftLength: 6,
		},
		CurrentIndex: 0,
		CurrentStart: time.Date(2018, 8, 30, 21, 54, 38, 334304000, time.UTC),
		Users: []string{
			"Javon Goodwin",
			"Nora Bode",
			"Coby Blanda",
			"Clyde Reinger",
			"Justina Moen",
			"Herman Donnelly",
			"Timmothy OReilly",
			"Alvis Upton",
			"Name Bayer",
			"Daron Hirthe",
			"Ismael Goodwin",
			"Andrew Lockman",
			"Adalberto Dare",
			"Liliana Moen",
			"Brant Abbott",
			"Nia Purdy",
			"Modesto Nolan",
			"Angelica Leannon",
			"Cleo Heaney",
			"Osborne Batz",
			"Lyda Christiansen",
			"Loyal Green",
			"Mose Lindgren",
			"Camilla Stehr",
		},
	}

	id := rot.UserID(time.Date(2018, 9, 10, 2, 44, 0, 0, time.UTC))
	if id != "Javon Goodwin" {
		t.Fatalf("got '%s'; want '%s'", id, "Javon Goodwin")
	}
}

func TestState_CalculateShifts(t *testing.T) {
	check := func(name string, start, end time.Time, s *state, exp []Shift) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			t.Helper()
			res := s.CalculateShifts(start, end)
			for i, exp := range exp {
				if i >= len(res) {
					t.Errorf("shift[%d]: missing", i)
					continue
				}
				if !res[i].Start.Equal(exp.Start) {
					t.Errorf("shift[%d]: start = %s; want %s", i, res[i].Start.In(exp.Start.Location()), exp.Start)
				}
				if !res[i].End.Equal(exp.End) {
					t.Errorf("shift[%d]: end = %s; want %s", i, res[i].End.In(exp.End.Location()), exp.End)
				}
				if res[i].Truncated != exp.Truncated {
					t.Errorf("shift[%d]: truncated = %t; want %t", i, res[i].Truncated, exp.Truncated)
				}
				if res[i].UserID != exp.UserID {
					t.Errorf("shift[%d]: userID = %s; want %s", i, res[i].UserID, exp.UserID)
				}
			}
			if len(res) > len(exp) {
				for _, res := range res[len(exp):] {
					t.Errorf("extra shift: %v", res)
				}
			}
		})
	}

	check("ActiveFuture",
		time.Date(2018, 1, 1, 8, 0, 0, 0, time.UTC), // 8:00AM
		time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC), // 9:00AM
		&state{
			loc: time.UTC,
			now: time.Date(2018, 1, 1, 7, 0, 0, 0, time.UTC),
			history: []Shift{
				{
					UserID: "still-active",
					Start:  time.Date(2018, 1, 1, 6, 0, 0, 0, time.UTC),
				},
				{
					UserID: "has-gap",
					Start:  time.Date(2018, 1, 1, 6, 0, 0, 0, time.UTC),
				},
			},
			rules: []ResolvedRule{
				{Rule: rule.Rule{
					WeekdayFilter: timeutil.WeekdayFilter{1, 1, 1, 1, 1, 1, 1},
					Start:         timeutil.NewClock(0, 0),
					End:           timeutil.NewClock(0, 0),
					Target:        assignment.UserTarget("still-active"),
				}},
				{Rule: rule.Rule{
					WeekdayFilter: timeutil.WeekdayFilter{0, 1, 0, 0, 0, 0, 0},
					Start:         timeutil.NewClock(8, 30),
					End:           timeutil.NewClock(10, 0),
					Target:        assignment.UserTarget("has-gap"),
				}},
			},
		},
		[]Shift{
			{
				Start:     time.Date(2018, 1, 1, 6, 0, 0, 0, time.UTC),
				End:       time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC),
				Truncated: true,
				UserID:    "still-active",
			},
			{
				Start:     time.Date(2018, 1, 1, 8, 30, 0, 0, time.UTC),
				End:       time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC),
				Truncated: true,
				UserID:    "has-gap",
			},
		},
	)

	check("HistoryRemainder",
		time.Date(2018, 1, 1, 8, 0, 0, 0, time.UTC), // 8:00AM
		time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC), // 9:00AM
		&state{
			now: time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC),
			loc: time.UTC,
			history: []Shift{
				{
					UserID: "foobar",
					Start:  time.Date(2018, 1, 1, 7, 0, 0, 0, time.UTC),
					End:    time.Date(2018, 1, 1, 8, 0, 0, 1, time.UTC), // will be truncated to 8
				},
			},
			rules: []ResolvedRule{
				{Rule: rule.Rule{
					WeekdayFilter: timeutil.WeekdayFilter{1, 1, 1, 1, 1, 1, 1},
					Start:         timeutil.NewClock(8, 0),
					End:           timeutil.NewClock(10, 0),
					Target:        assignment.UserTarget("foobar"),
				}},
			},
		},
		[]Shift{
			// no shift is expected since it ended before/at the start time
		},
	)

	check("Simple",
		time.Date(2018, 1, 1, 8, 0, 0, 0, time.UTC), // 8:00AM
		time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC), // 9:00AM
		&state{
			loc: time.UTC,
			rules: []ResolvedRule{
				{Rule: rule.Rule{
					WeekdayFilter: timeutil.WeekdayFilter{1, 1, 1, 1, 1, 1, 1},
					Start:         timeutil.NewClock(8, 0),
					End:           timeutil.NewClock(10, 0),
					Target:        assignment.UserTarget("foobar"),
				}},
			},
		},
		[]Shift{
			{
				Start:     time.Date(2018, 1, 1, 8, 0, 0, 0, time.UTC),
				End:       time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC),
				Truncated: true,
				UserID:    "foobar",
			},
		},
	)

	check("Temporary Schedule",
		time.Date(2018, 1, 1, 8, 0, 0, 0, time.UTC), // 8:00AM
		time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC), // 9:00AM
		&state{
			loc: time.UTC,
			rules: []ResolvedRule{
				{Rule: rule.Rule{
					WeekdayFilter: timeutil.WeekdayFilter{1, 1, 1, 1, 1, 1, 1},
					Start:         timeutil.NewClock(8, 0),
					End:           timeutil.NewClock(10, 0),
					Target:        assignment.UserTarget("foobar"),
				}},
			},
			tempScheds: []schedule.TemporarySchedule{
				{
					Start: time.Date(2018, 1, 1, 8, 15, 0, 0, time.UTC),
					End:   time.Date(2018, 1, 1, 8, 45, 0, 0, time.UTC),
					Shifts: []schedule.FixedShift{{
						Start:  time.Date(2018, 1, 1, 8, 25, 0, 0, time.UTC),
						End:    time.Date(2018, 1, 1, 8, 35, 0, 0, time.UTC),
						UserID: "baz",
					}},
				},
			},
		},
		[]Shift{
			{
				Start:     time.Date(2018, 1, 1, 8, 0, 0, 0, time.UTC),
				End:       time.Date(2018, 1, 1, 8, 15, 0, 0, time.UTC),
				Truncated: false,
				UserID:    "foobar",
			},
			{
				Start:     time.Date(2018, 1, 1, 8, 25, 0, 0, time.UTC),
				End:       time.Date(2018, 1, 1, 8, 35, 0, 0, time.UTC),
				Truncated: false,
				UserID:    "baz",
			},
			{
				Start:     time.Date(2018, 1, 1, 8, 45, 0, 0, time.UTC),
				End:       time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC),
				Truncated: true,
				UserID:    "foobar",
			},
		},
	)

	check("SimpleWeek",
		time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2018, 1, 8, 0, 0, 0, 0, time.UTC),
		&state{
			loc: time.UTC,
			rules: []ResolvedRule{
				{Rule: rule.Rule{
					WeekdayFilter: timeutil.WeekdayFilter{1, 1, 1, 1, 1, 1, 1},
					Start:         timeutil.NewClock(8, 0),
					End:           timeutil.NewClock(9, 0),
					Target:        assignment.UserTarget("foobar"),
				}},
			},
		},
		[]Shift{
			{
				Start:  time.Date(2018, 1, 1, 8, 0, 0, 0, time.UTC),
				End:    time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC),
				UserID: "foobar",
			},
			{
				Start:  time.Date(2018, 1, 2, 8, 0, 0, 0, time.UTC),
				End:    time.Date(2018, 1, 2, 9, 0, 0, 0, time.UTC),
				UserID: "foobar",
			},
			{
				Start:  time.Date(2018, 1, 3, 8, 0, 0, 0, time.UTC),
				End:    time.Date(2018, 1, 3, 9, 0, 0, 0, time.UTC),
				UserID: "foobar",
			},
			{
				Start:  time.Date(2018, 1, 4, 8, 0, 0, 0, time.UTC),
				End:    time.Date(2018, 1, 4, 9, 0, 0, 0, time.UTC),
				UserID: "foobar",
			},
			{
				Start:  time.Date(2018, 1, 5, 8, 0, 0, 0, time.UTC),
				End:    time.Date(2018, 1, 5, 9, 0, 0, 0, time.UTC),
				UserID: "foobar",
			},
			{
				Start:  time.Date(2018, 1, 6, 8, 0, 0, 0, time.UTC),
				End:    time.Date(2018, 1, 6, 9, 0, 0, 0, time.UTC),
				UserID: "foobar",
			},
			{
				Start:  time.Date(2018, 1, 7, 8, 0, 0, 0, time.UTC),
				End:    time.Date(2018, 1, 7, 9, 0, 0, 0, time.UTC),
				UserID: "foobar",
			},
		},
	)

	check("ReplaceOverride",
		time.Date(2018, 1, 1, 8, 0, 0, 0, time.UTC), // 8:00AM
		time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC), // 9:00AM
		&state{
			loc: time.UTC,
			rules: []ResolvedRule{
				{Rule: rule.Rule{
					WeekdayFilter: timeutil.WeekdayFilter{1, 1, 1, 1, 1, 1, 1},
					Start:         timeutil.NewClock(8, 0),
					End:           timeutil.NewClock(10, 0),
					Target:        assignment.UserTarget("foobar"),
				}},
			},
			overrides: []override.UserOverride{
				{
					AddUserID:    "binbaz",
					RemoveUserID: "foobar",
					Start:        time.Date(2018, 1, 1, 8, 30, 0, 0, time.UTC),
					End:          time.Date(2018, 1, 1, 8, 45, 0, 0, time.UTC),
				},
			},
		},
		[]Shift{
			{
				Start:     time.Date(2018, 1, 1, 8, 0, 0, 0, time.UTC),
				End:       time.Date(2018, 1, 1, 8, 30, 0, 0, time.UTC),
				Truncated: false,
				UserID:    "foobar",
			},
			{
				Start:     time.Date(2018, 1, 1, 8, 30, 0, 0, time.UTC),
				End:       time.Date(2018, 1, 1, 8, 45, 0, 0, time.UTC),
				Truncated: false,
				UserID:    "binbaz",
			},
			{
				Start:     time.Date(2018, 1, 1, 8, 45, 0, 0, time.UTC),
				End:       time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC),
				Truncated: true,
				UserID:    "foobar",
			},
		},
	)

	check("AddOverride",
		time.Date(2018, 1, 1, 8, 0, 0, 0, time.UTC), // 8:00AM
		time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC), // 9:00AM
		&state{
			loc: time.UTC,
			rules: []ResolvedRule{
				{Rule: rule.Rule{
					WeekdayFilter: timeutil.WeekdayFilter{1, 1, 1, 1, 1, 1, 1},
					Start:         timeutil.NewClock(8, 0),
					End:           timeutil.NewClock(10, 0),
					Target:        assignment.UserTarget("foobar"),
				}},
			},
			overrides: []override.UserOverride{
				{
					AddUserID: "binbaz",
					Start:     time.Date(2018, 1, 1, 8, 30, 0, 0, time.UTC),
					End:       time.Date(2018, 1, 1, 8, 45, 0, 0, time.UTC),
				},

				{
					AddUserID: "binbaz",
					Start:     time.Date(2018, 1, 1, 8, 0, 0, 0, time.UTC),
					End:       time.Date(2018, 1, 1, 8, 15, 0, 0, time.UTC),
				},
			},
		},
		[]Shift{
			{
				Start:     time.Date(2018, 1, 1, 8, 0, 0, 0, time.UTC),
				End:       time.Date(2018, 1, 1, 8, 15, 0, 0, time.UTC),
				Truncated: false,
				UserID:    "binbaz",
			},
			{
				Start:     time.Date(2018, 1, 1, 8, 0, 0, 0, time.UTC),
				End:       time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC),
				Truncated: true,
				UserID:    "foobar",
			},
			{
				Start:     time.Date(2018, 1, 1, 8, 30, 0, 0, time.UTC),
				End:       time.Date(2018, 1, 1, 8, 45, 0, 0, time.UTC),
				Truncated: false,
				UserID:    "binbaz",
			},
		},
	)

	check("RemoveOverride",
		time.Date(2018, 1, 1, 8, 0, 0, 0, time.UTC), // 8:00AM
		time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC), // 9:00AM
		&state{
			loc: time.UTC,
			rules: []ResolvedRule{
				{Rule: rule.Rule{
					WeekdayFilter: timeutil.WeekdayFilter{1, 1, 1, 1, 1, 1, 1},
					Start:         timeutil.NewClock(8, 0),
					End:           timeutil.NewClock(10, 0),
					Target:        assignment.UserTarget("foobar"),
				}},
			},
			overrides: []override.UserOverride{
				{
					RemoveUserID: "foobar",
					Start:        time.Date(2018, 1, 1, 8, 30, 0, 0, time.UTC),
					End:          time.Date(2018, 1, 1, 8, 45, 0, 0, time.UTC),
				},
			},
		},
		[]Shift{
			{
				Start:     time.Date(2018, 1, 1, 8, 0, 0, 0, time.UTC),
				End:       time.Date(2018, 1, 1, 8, 30, 0, 0, time.UTC),
				Truncated: false,
				UserID:    "foobar",
			},
			{
				Start:     time.Date(2018, 1, 1, 8, 45, 0, 0, time.UTC),
				End:       time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC),
				Truncated: true,
				UserID:    "foobar",
			},
		},
	)

	check("History",
		time.Date(2018, 1, 1, 8, 0, 0, 0, time.UTC), // 8:00AM
		time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC), // 9:00AM
		&state{
			now: time.Date(2018, 1, 1, 8, 0, 0, 0, time.UTC),
			loc: time.UTC,
			history: []Shift{
				{
					UserID: "foobar",
					Start:  time.Date(2018, 1, 1, 7, 0, 0, 0, time.UTC), // user actually started at 7
				},
			},
			rules: []ResolvedRule{
				{Rule: rule.Rule{
					WeekdayFilter: timeutil.WeekdayFilter{1, 1, 1, 1, 1, 1, 1},
					Start:         timeutil.NewClock(8, 0),
					End:           timeutil.NewClock(10, 0),
					Target:        assignment.UserTarget("foobar"),
				}},
			},
		},
		[]Shift{
			{
				Start:     time.Date(2018, 1, 1, 7, 0, 0, 0, time.UTC),
				End:       time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC),
				Truncated: true,
				UserID:    "foobar",
			},
		},
	)

	check("Rotation",
		time.Date(2018, 9, 10, 0, 0, 0, 0, time.UTC), // 8:00AM
		time.Date(2018, 9, 17, 0, 0, 0, 0, time.UTC), // 9:00AM
		&state{
			now: time.Date(2018, 9, 10, 14, 44, 0, 0, time.UTC),
			loc: time.UTC,
			history: []Shift{
				{
					UserID: "Javon Goodwin",
					Start:  time.Date(2018, 9, 10, 2, 25, 0, 0, time.UTC),
					End:    time.Date(2018, 9, 10, 5, 29, 0, 0, time.UTC),
				},
			},
			rules: []ResolvedRule{
				{Rule: rule.Rule{
					WeekdayFilter: timeutil.WeekdayFilter{1, 1, 1, 1, 0, 0, 0},
					Start:         timeutil.NewClock(19, 40),
					End:           timeutil.NewClock(22, 53),
					Target:        assignment.RotationTarget("rot"),
				},
					Rotation: &ResolvedRotation{
						Rotation: rotation.Rotation{
							ID:          "rot",
							Type:        rotation.TypeWeekly,
							Start:       time.Date(2018, 4, 21, 14, 8, 9, 168379000, time.UTC),
							ShiftLength: 6,
						},
						CurrentIndex: 0,
						CurrentStart: time.Date(2018, 8, 30, 21, 54, 38, 334304000, time.UTC),
						Users: []string{
							"Javon Goodwin",
							"Nora Bode",
							"Coby Blanda",
							"Clyde Reinger",
							"Justina Moen",
							"Herman Donnelly",
							"Timmothy OReilly",
							"Alvis Upton",
							"Name Bayer",
							"Daron Hirthe",
							"Ismael Goodwin",
							"Andrew Lockman",
							"Adalberto Dare",
							"Liliana Moen",
							"Brant Abbott",
							"Nia Purdy",
							"Modesto Nolan",
							"Angelica Leannon",
							"Cleo Heaney",
							"Osborne Batz",
							"Lyda Christiansen",
							"Loyal Green",
							"Mose Lindgren",
							"Camilla Stehr",
						},
					},
				},
			},
		},
		[]Shift{
			{
				Start:     time.Date(2018, 9, 10, 2, 25, 0, 0, time.UTC),
				End:       time.Date(2018, 9, 10, 5, 29, 0, 0, time.UTC),
				Truncated: false,
				UserID:    "Javon Goodwin",
			},
			{
				Start:     time.Date(2018, 9, 10, 19, 40, 0, 0, time.UTC),
				End:       time.Date(2018, 9, 10, 22, 53, 0, 0, time.UTC),
				Truncated: false,
				UserID:    "Javon Goodwin",
			},
			{
				Start:     time.Date(2018, 9, 11, 19, 40, 0, 0, time.UTC),
				End:       time.Date(2018, 9, 11, 22, 53, 0, 0, time.UTC),
				Truncated: false,
				UserID:    "Javon Goodwin",
			},
			{
				Start:     time.Date(2018, 9, 12, 19, 40, 0, 0, time.UTC),
				End:       time.Date(2018, 9, 12, 22, 53, 0, 0, time.UTC),
				Truncated: false,
				UserID:    "Javon Goodwin",
			},
			{
				Start:     time.Date(2018, 9, 16, 19, 40, 0, 0, time.UTC),
				End:       time.Date(2018, 9, 16, 22, 53, 0, 0, time.UTC),
				Truncated: false,
				UserID:    "Javon Goodwin",
			},
		},
	)

	central, err := time.LoadLocation("America/Chicago")
	if err != nil {
		t.Fatal(err)
	}
	check(
		"DailyRotation",
		time.Date(2018, 9, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2018, 9, 17, 0, 0, 0, 0, time.UTC),
		&state{
			now: time.Date(2018, 9, 12, 10, 25, 0, 0, central),
			loc: central,
			history: []Shift{
				{UserID: "Craig", Start: time.Date(2018, 9, 12, 13, 5, 0, 0, time.UTC)},
				{UserID: "Caza", Start: time.Date(2018, 9, 12, 13, 0, 3, 0, time.UTC), End: time.Date(2018, 9, 12, 13, 1, 3, 0, time.UTC)},
				{UserID: "Cook", Start: time.Date(2018, 9, 12, 1, 0, 3, 0, time.UTC), End: time.Date(2018, 9, 12, 13, 0, 3, 0, time.UTC)},
				{UserID: "Aru", Start: time.Date(2018, 9, 11, 13, 0, 1, 0, time.UTC), End: time.Date(2018, 9, 12, 1, 1, 3, 0, time.UTC)},
				{UserID: "Caza", Start: time.Date(2018, 9, 11, 1, 0, 1, 0, time.UTC), End: time.Date(2018, 9, 11, 13, 0, 1, 0, time.UTC)},
				{UserID: "Donna", Start: time.Date(2018, 9, 10, 13, 0, 0, 0, time.UTC), End: time.Date(2018, 9, 11, 1, 0, 1, 0, time.UTC)},
				{UserID: "Cook", Start: time.Date(2018, 9, 10, 1, 0, 0, 0, time.UTC), End: time.Date(2018, 9, 10, 13, 0, 0, 0, time.UTC)},
				{UserID: "Craig", Start: time.Date(2018, 9, 9, 13, 0, 0, 0, time.UTC), End: time.Date(2018, 9, 10, 1, 0, 0, 0, time.UTC)},
			},
			rules: []ResolvedRule{
				{Rule: rule.Rule{
					WeekdayFilter: timeutil.WeekdayFilter{1, 1, 1, 1, 1, 1, 1},
					Start:         timeutil.NewClock(8, 0),
					End:           timeutil.NewClock(20, 0),
					Target:        assignment.RotationTarget("rot-day"),
				},
					Rotation: &ResolvedRotation{
						Rotation: rotation.Rotation{
							ID:          "rot-day",
							Type:        rotation.TypeDaily,
							Start:       time.Date(2018, 6, 15, 13, 0, 0, 0, time.UTC).In(central),
							ShiftLength: 1,
						},
						CurrentIndex: 2,
						CurrentStart: time.Date(2018, 9, 12, 13, 0, 3, 0, time.UTC),
						Users: []string{
							"Donna",
							"Aru",
							"Craig",
						},
					},
				},
				{Rule: rule.Rule{
					WeekdayFilter: timeutil.WeekdayFilter{1, 1, 1, 1, 1, 1, 1},
					Start:         timeutil.NewClock(20, 0),
					End:           timeutil.NewClock(8, 0),
					Target:        assignment.RotationTarget("rot-night"),
				},
					Rotation: &ResolvedRotation{
						Rotation: rotation.Rotation{
							ID:          "rot-night",
							Type:        rotation.TypeDaily,
							Start:       time.Date(2018, 6, 15, 13, 0, 0, 0, time.UTC),
							ShiftLength: 1,
						},
						CurrentIndex: 0,
						CurrentStart: time.Date(2018, 9, 12, 13, 0, 3, 0, time.UTC),
						Users: []string{
							"Caza",
							"Cook",
						},
					},
				},
			},
		},
		[]Shift{
			// history shifts
			{UserID: "Craig", Start: time.Date(2018, 9, 9, 13, 0, 0, 0, time.UTC), End: time.Date(2018, 9, 10, 1, 0, 0, 0, time.UTC)},
			{UserID: "Cook", Start: time.Date(2018, 9, 10, 1, 0, 0, 0, time.UTC), End: time.Date(2018, 9, 10, 13, 0, 0, 0, time.UTC)},
			{UserID: "Donna", Start: time.Date(2018, 9, 10, 13, 0, 0, 0, time.UTC), End: time.Date(2018, 9, 11, 1, 0, 0, 0, time.UTC)},
			{UserID: "Caza", Start: time.Date(2018, 9, 11, 1, 0, 0, 0, time.UTC), End: time.Date(2018, 9, 11, 13, 0, 0, 0, time.UTC)},
			{UserID: "Aru", Start: time.Date(2018, 9, 11, 13, 0, 0, 0, time.UTC), End: time.Date(2018, 9, 12, 1, 1, 0, 0, time.UTC)},
			{UserID: "Cook", Start: time.Date(2018, 9, 12, 1, 0, 0, 0, time.UTC), End: time.Date(2018, 9, 12, 13, 0, 0, 0, time.UTC)},
			{UserID: "Caza", Start: time.Date(2018, 9, 12, 13, 0, 0, 0, time.UTC), End: time.Date(2018, 9, 12, 13, 1, 0, 0, time.UTC)},

			// in-progress
			{UserID: "Craig", Start: time.Date(2018, 9, 12, 13, 5, 0, 0, time.UTC), End: time.Date(2018, 9, 12, 20, 0, 0, 0, central)},

			// future
			{UserID: "Caza", Start: time.Date(2018, 9, 12, 20, 0, 0, 0, central), End: time.Date(2018, 9, 13, 8, 0, 0, 0, central)},
			{UserID: "Donna", Start: time.Date(2018, 9, 13, 8, 0, 0, 0, central), End: time.Date(2018, 9, 13, 20, 0, 0, 0, central)},
			{UserID: "Cook", Start: time.Date(2018, 9, 13, 20, 0, 0, 0, central), End: time.Date(2018, 9, 14, 8, 0, 0, 0, central)},
			{UserID: "Aru", Start: time.Date(2018, 9, 14, 8, 0, 0, 0, central), End: time.Date(2018, 9, 14, 20, 0, 0, 0, central)},
			{UserID: "Caza", Start: time.Date(2018, 9, 14, 20, 0, 0, 0, central), End: time.Date(2018, 9, 15, 8, 0, 0, 0, central)},
			{UserID: "Craig", Start: time.Date(2018, 9, 15, 8, 0, 0, 0, central), End: time.Date(2018, 9, 15, 20, 0, 0, 0, central)},
			{UserID: "Cook", Start: time.Date(2018, 9, 15, 20, 0, 0, 0, central), End: time.Date(2018, 9, 16, 8, 0, 0, 0, central)},
			{UserID: "Donna", Start: time.Date(2018, 9, 16, 8, 0, 0, 0, central), End: time.Date(2018, 9, 16, 19, 0, 0, 0, central), Truncated: true},
		},
	)

}
