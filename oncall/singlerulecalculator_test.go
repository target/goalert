package oncall_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/util/timeutil"
)

func TestSingleRuleCalculator(t *testing.T) {
	type result struct {
		Time  time.Time
		Value string
	}
	var (
		start = time.Date(2000, 1, 2, 3, 4, 0, 0, time.UTC)
		end   = time.Date(2000, 1, 2, 3, 8, 0, 0, time.UTC)
	)
	check := func(desc string, expected []result, rule oncall.ResolvedRule) {
		t.Run(desc, func(t *testing.T) {
			iter := oncall.NewTimeIterator(
				start,
				end,
				time.Minute,
			).NewSingleRuleCalculator(time.UTC, rule)

			var results []result
			for iter.Next() {
				results = append(results, result{Time: time.Unix(iter.Unix(), 0).UTC(), Value: iter.ActiveUser()})
			}

			assert.ElementsMatch(t, expected, results)
		})
	}
	check("empty",
		[]result{
			{Time: start},
			{Time: end},
		},
		oncall.ResolvedRule{},
	)

	check("simple",
		[]result{
			{Time: start},
			{Time: time.Date(2000, 1, 2, 3, 5, 0, 0, time.UTC), Value: "foo"},
			{Time: time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC)},
			{Time: end},
		},
		oncall.ResolvedRule{
			Rule: rule.Rule{
				Start:         timeutil.NewClock(3, 5),
				End:           timeutil.NewClock(3, 7),
				WeekdayFilter: timeutil.EveryDay(),
				Target:        assignment.UserTarget("foo"),
			},
		},
	)
}

func TestDailyRotationWithDayRestrictions(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Amsterdam")

	t.Run("SmallRotation", func(t *testing.T) {
		userIDs := []string{"user-0", "user-1", "user-2", "user-3", "user-4", "user-5", "user-6", "user-7"}
		rotStart := time.Date(2026, 2, 1, 9, 0, 0, 0, loc)

		rot := &oncall.ResolvedRotation{
			Rotation: rotation.Rotation{
				ID:          "test-rotation",
				Type:        rotation.TypeDaily,
				ShiftLength: 1,
				Start:       rotStart,
			},
			CurrentIndex: 0,
			CurrentStart: rotStart,
			Users:        userIDs,
		}

		testRule := oncall.ResolvedRule{
			Rule: rule.Rule{
				WeekdayFilter: timeutil.WeekdayFilter{0, 1, 1, 1, 1, 1, 0},
				Start:         timeutil.NewClock(9, 0),
				End:           timeutil.NewClock(17, 0),
			},
			Rotation: rot,
		}

		start := time.Date(2026, 2, 1, 0, 0, 0, 0, loc)
		end := time.Date(2026, 2, 15, 0, 0, 0, 0, loc)

		iter := oncall.NewTimeIterator(start, end, time.Minute).NewSingleRuleCalculator(loc, testRule)

		shiftsByUser := make(map[string]int)
		var currentUser string
		for iter.Next() {
			user := iter.ActiveUser()
			if user != "" && user != currentUser {
				shiftsByUser[user]++
				currentUser = user
			}
		}

		if len(shiftsByUser) != 8 {
			t.Errorf("Expected 8 unique users, got %d", len(shiftsByUser))
			for _, userID := range userIDs {
				if shiftsByUser[userID] == 0 {
					t.Errorf("  Missing: %s", userID)
				}
			}
		}
	})

	t.Run("LargeRotation", func(t *testing.T) {
		userIDs := []string{
			"user-0", "user-1", "user-2", "user-3", "user-4", "user-5", "user-6",
			"user-7", "user-8", "user-9", "user-10", "user-11", "user-12", "user-13",
		}

		loc, _ := time.LoadLocation("Europe/Amsterdam")
		rotStart := time.Date(2025, 12, 3, 9, 0, 0, 0, loc)

		rot := &oncall.ResolvedRotation{
			Rotation: rotation.Rotation{
				ID:          "test-rotation",
				Type:        rotation.TypeDaily,
				ShiftLength: 1,
				Start:       rotStart,
			},
			CurrentIndex: 5,
			CurrentStart: time.Date(2025, 12, 8, 9, 0, 0, 0, loc),
			Users:        userIDs,
		}

		testRule := oncall.ResolvedRule{
			Rule: rule.Rule{
				WeekdayFilter: timeutil.WeekdayFilter{0, 1, 1, 1, 1, 1, 0},
				Start:         timeutil.NewClock(9, 0),
				End:           timeutil.NewClock(17, 0),
			},
			Rotation: rot,
		}

		start := time.Date(2026, 2, 1, 0, 0, 0, 0, loc)
		end := time.Date(2026, 3, 1, 0, 0, 0, 0, loc)

		iter := oncall.NewTimeIterator(start, end, time.Minute).NewSingleRuleCalculator(loc, testRule)

		shiftsByUser := make(map[string]int)
		var currentUser string
		for iter.Next() {
			user := iter.ActiveUser()
			if user != "" && user != currentUser {
				shiftsByUser[user]++
				currentUser = user
			}
		}

		if len(shiftsByUser) != 14 {
			t.Errorf("Expected 14 unique users, got %d", len(shiftsByUser))
			for _, userID := range userIDs {
				if shiftsByUser[userID] == 0 {
					t.Errorf("  Missing: %s (never appeared)", userID)
				}
			}
		}
	})
}
