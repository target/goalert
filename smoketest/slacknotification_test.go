package smoketest

import (
	"testing"
	"time"

	"github.com/target/goalert/smoketest/harness"
)

// TestSlackNotification tests that slack channels are returned for configured users.
func TestSlackNotification(t *testing.T) {
	t.Parallel()

	sql := `
	insert into escalation_policies (id, name, repeat) 
	values
		({{uuid "eid"}}, 'esc policy', 1);
	insert into escalation_policy_steps (id, escalation_policy_id, delay) 
	values
		({{uuid "esid"}}, {{uuid "eid"}}, 30);

	insert into notification_channels (id, type, name, value)
	values
		({{uuid "chan"}}, 'SLACK', '#test', {{slackChannelID "test"}});

	insert into escalation_policy_actions (escalation_policy_step_id, channel_id) 
	values 
		({{uuid "esid"}}, {{uuid "chan"}});

	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');
`
	h := harness.NewHarness(t, sql, "slack-user-link")
	defer h.Close()

	h.CreateAlert(h.UUID("sid"), "testing")
	msg := h.Slack().Channel("test").ExpectMessage("testing")

	h.FastForward(time.Hour)
	// should broadcast reply to channel
	msg.ExpectBroadcastReply("repeat notification")
}
