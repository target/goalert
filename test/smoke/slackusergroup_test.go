package smoke

import (
	"testing"
	"time"

	"github.com/target/goalert/test/smoke/harness"
)

// TestSlackUserGroups tests that the configured notification rule sends the intended notification to the slack user group.
func TestSlackUserGroups(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email)
	values
		({{uuid "uid"}}, 'bob', 'bob@example.com');

	insert into user_contact_methods (id, user_id, name, type, value, pending)
	values
		({{uuid "cm1"}}, {{uuid "uid"}}, 'personal', 'SLACK_DM', {{slackUserID "bob"}}, false);

	insert into schedules (id, name, time_zone) 
	values
		({{uuid "sid"}}, 'testschedule', 'UTC');

	insert into schedule_rules (id, schedule_id, sunday, monday, tuesday, wednesday, thursday, friday, saturday, start_time, end_time, tgt_user_id)
	values
		({{uuid "ruleID"}}, {{uuid "sid"}}, true, true, true, true, true, true, true, '00:00:00', '00:00:00', {{uuid "uid"}});

	insert into notification_channels (id, type, name, value)
	values
		({{uuid "ug"}}, 'SLACK_USER_GROUP', '@testug (#test1)', {{slackUserGroupID "test2"}});
	
	insert into schedule_data (schedule_id, data)
	values
		({{uuid "sid"}}, '{"V1":{"OnCallNotificationRules": [{"ChannelID": {{uuidJSON "ug"}}, "Time": "00:00" }]}}');
`
	h := harness.NewHarness(t, sql, "slack-ug")

	defer h.Close()

	h.Trigger()

	// Passing in no arguments to assert empty
	h.Slack().UserGroup("test2").ExpectUsers()

	h.FastForward(24 * time.Hour)

	h.Slack().UserGroup("test2").ExpectUsers("bob")
}
