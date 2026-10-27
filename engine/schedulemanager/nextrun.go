package schedulemanager

import (
	"time"

	"github.com/target/goalert/gadb"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/util/timeutil"
)

// ruleFromRow converts a schedule-manager rule row into a rule.Rule carrying
// only the fields relevant to time math (start, end, weekday filter).
func ruleFromRow(row gadb.SchedMgrRulesRow) rule.Rule {
	var wf timeutil.WeekdayFilter
	wf.SetDay(time.Sunday, row.Sunday)
	wf.SetDay(time.Monday, row.Monday)
	wf.SetDay(time.Tuesday, row.Tuesday)
	wf.SetDay(time.Wednesday, row.Wednesday)
	wf.SetDay(time.Thursday, row.Thursday)
	wf.SetDay(time.Friday, row.Friday)
	wf.SetDay(time.Saturday, row.Saturday)

	return rule.Rule{
		Start:         row.StartTime,
		End:           row.EndTime,
		WeekdayFilter: wf,
	}
}
