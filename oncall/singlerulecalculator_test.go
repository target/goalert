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
	check := func(desc string, results []string, rule oncall.ResolvedRule) {
		t.Run(desc, func(t *testing.T) {
			iter := oncall.NewTimeIterator(
				time.Date(2000, 1, 2, 3, 4, 0, 0, time.UTC),
				time.Date(2000, 1, 2, 3, 8, 0, 0, time.UTC),
				time.Minute,
			).NewSingleRuleCalculator(time.UTC, rule)

			var last string
			for i, exp := range results {
				assert.Truef(t, iter.Next(), "Next() call #%d", i+1)
				assert.Equalf(t, last != exp, iter.Changed(), "Changed() call #%d", i+1)
				assert.EqualValuesf(t, exp, iter.ActiveUser(), "Active() call #%d", i+1)
				last = exp
			}

			assert.Falsef(t, iter.Next(), "Next() call #%d (last)", len(results)+1)
		})
	}
	check("empty", []string{"", "", "", "", ""}, oncall.ResolvedRule{})

	check("simple", []string{"", "foo", "foo", "", ""}, oncall.ResolvedRule{
		Rule: rule.Rule{
			Start:         rule.NewClock(3, 5),
			End:           rule.NewClock(3, 7),
			WeekdayFilter: rule.EveryDay(),
			Target:        assignment.UserTarget("foo"),
		},
	})
}
