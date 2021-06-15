package smoketest

import (
	"testing"

	"github.com/target/goalert/smoketest/harness"
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
}
