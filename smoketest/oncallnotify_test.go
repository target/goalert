package smoketest

import (
	"context"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/smoketest/harness"
	"github.com/target/goalert/util/timeutil"
)

func TestOnCallNotify(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email)
	values
		({{uuid "uid"}}, 'bob', 'bob@example.com'),
		({{uuid "uid2"}}, 'joe', 'joe@example.com');

	insert into schedules (id, name, time_zone) 
	values
		({{uuid "sid"}}, 'testschedule', 'UTC');

	insert into schedule_rules (id, schedule_id, sunday, monday, tuesday, wednesday, thursday, friday, saturday, start_time, end_time, tgt_user_id)
	values
		({{uuid "ruleID"}}, {{uuid "sid"}}, true, true, true, true, true, true, true, '00:00:00', '00:00:00', {{uuid "uid"}});

	insert into notification_channels (id, type, name, value)
	values
		({{uuid "chan"}}, 'SLACK', '#test', {{slackChannelID "test"}});
	
	insert into schedule_data (schedule_id, data)
	values
		({{uuid "sid"}}, '{"V1":{"OnCallNotificationRules": [{"ChannelID": {{uuidJSON "chan"}} }]}}');
`
	h := harness.NewHarness(t, sql, "outgoing-messages-schedule-id")
	defer h.Close()

	h.Slack().Channel("test").ExpectMessage("on-call", "testschedule", "bob")

	ctx := permission.SystemContext(context.Background(), "Test")
	_, err := h.App().ScheduleRuleStore.CreateRuleTx(ctx, nil, &rule.Rule{
		ID:            uuid.NewV4().String(),
		ScheduleID:    h.UUID("sid"),
		WeekdayFilter: timeutil.EveryDay(),
		Target:        assignment.UserTarget(h.UUID("uid2")),
	})
	require.NoError(t, err)

	h.Slack().Channel("test").ExpectMessage("on-call", "testschedule", "bob", "joe")
}
