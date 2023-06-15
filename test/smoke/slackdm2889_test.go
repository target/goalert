package smoke

import (
	"testing"

	"github.com/target/goalert/test/smoke/harness"
)

// TestSlackDM2889 tests that slack DMs update correctly, even when the old status log config is present (issue #2889).
func TestSlackDM2889(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email, role) 
	values 
		({{uuid "user"}}, 'bob', 'joe', 'user');
	insert into user_contact_methods (id, user_id, name, type, value)
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SLACK_DM', {{slackUserID "bob"}}),
		({{uuid "cm2"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "bob"}});
	
	update users set alert_status_log_contact_method_id = {{uuid "cm2"}} where id = {{uuid "user"}};

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "user"}}, {{uuid "cm1"}}, 0),
		({{uuid "user"}}, {{uuid "cm1"}}, 30);

	insert into escalation_policies (id, name) 
	values
		({{uuid "eid"}}, 'esc policy');
	insert into escalation_policy_steps (id, escalation_policy_id) 
	values
		({{uuid "esid"}}, {{uuid "eid"}});
	insert into escalation_policy_actions (escalation_policy_step_id, user_id) 
	values 
		({{uuid "esid"}}, {{uuid "user"}});

	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');

`
	h := harness.NewHarness(t, sql, "slack-dm-cm-type")
	defer h.Close()
	h.SetConfigValue("Slack.InteractiveMessages", "true")
	h.Trigger() // the user's account should get "linked" via compat mgr

	h.CreateAlert(h.UUID("sid"), "testing")
	msg := h.Slack().User("bob").ExpectMessage("testing")
	msg.Action("Close").Click()

	updated := msg.ExpectUpdate()
	updated.AssertText("Closed", "testing")
	updated.AssertActions()
}
