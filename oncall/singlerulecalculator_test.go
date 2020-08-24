package oncall_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/schedule/rule"
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

			assert.EqualValues(t, expected, results)
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
				Start:         rule.NewClock(3, 5),
				End:           rule.NewClock(3, 7),
				WeekdayFilter: rule.EveryDay(),
				Target:        assignment.UserTarget("foo"),
			},
		},
	)
}
