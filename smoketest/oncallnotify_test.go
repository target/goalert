package smoketest

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/smoketest/harness"
	"github.com/target/goalert/util/timeutil"
)

// TestOnCallNotify will validate that on-change notifications are sent for schedules.
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
		({{uuid "chan1"}}, 'SLACK', '#test1', {{slackChannelID "test1"}}),
		({{uuid "chan2"}}, 'SLACK', '#test2', {{slackChannelID "test2"}});

	insert into schedule_data (schedule_id, data)
	values
		({{uuid "sid"}}, '{"V1":{"OnCallNotificationRules": [{"ChannelID": {{uuidJSON "chan1"}} }, {"ChannelID": {{uuidJSON "chan2"}} }]}}');
`
	h := harness.NewHarness(t, sql, "outgoing-messages-schedule-id")
	defer h.Close()

	h.Slack().Channel("test").ExpectMessage("on-call", "testschedule", "bob")

	h.FastForward(time.Hour)

	ctx := permission.SystemContext(context.Background(), "Test")
	_, err := h.App().ScheduleRuleStore.CreateRuleTx(ctx, nil, &rule.Rule{
		ID:            uuid.New().String(),
		ScheduleID:    h.UUID("sid"),
		WeekdayFilter: timeutil.EveryDay(),
		Target:        assignment.UserTarget(h.UUID("uid2")),
	})
	require.NoError(t, err)

	h.Slack().Channel("test1").ExpectMessage("on-call", "testschedule", "bob", "joe")
	h.Slack().Channel("test2").ExpectMessage("on-call", "testschedule", "bob", "joe")
}
