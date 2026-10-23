package schedulemanager

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/util/timeutil"
)

func TestRuleFromRow(t *testing.T) {
	row := gadb.SchedMgrRulesRow{
		StartTime: timeutil.NewClock(8, 0),
		EndTime:   timeutil.NewClock(20, 0),
		Monday:    true,
		Wednesday: true,
		Friday:    true,
	}
	assert.Equal(t, rule.Rule{
		Start:         timeutil.NewClock(8, 0),
		End:           timeutil.NewClock(20, 0),
		WeekdayFilter: timeutil.WeekdayFilter{0, 1, 0, 1, 0, 1, 0},
	}, ruleFromRow(row))

	row = gadb.SchedMgrRulesRow{
		StartTime: timeutil.NewClock(22, 15),
		EndTime:   timeutil.NewClock(6, 45),
		Sunday:    true,
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		Saturday:  true,
	}
	assert.Equal(t, rule.Rule{
		Start:         timeutil.NewClock(22, 15),
		End:           timeutil.NewClock(6, 45),
		WeekdayFilter: timeutil.EveryDay(),
	}, ruleFromRow(row))

	// ruleRowIsActive must remain equivalent to the extracted rule
	tm := time.Date(2026, 10, 21, 23, 0, 0, 0, time.UTC) // Wednesday
	assert.True(t, ruleRowIsActive(row, tm))
	assert.Equal(t, ruleFromRow(row).IsActive(tm), ruleRowIsActive(row, tm))
}
