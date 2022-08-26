package smoke

import (
	"testing"

	"github.com/target/goalert/test/smoke/harness"
)

// TestStatusUpdatesChannel tests status updates to notification channels.
func TestStatusUpdatesChannel(t *testing.T) {
	t.Parallel()

	sql := `
	insert into escalation_policies (id, name) 
	values
		({{uuid "eid"}}, 'esc policy');
	insert into escalation_policy_steps (id, escalation_policy_id) 
	values
		({{uuid "esid"}}, {{uuid "eid"}});

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

	a := h.CreateAlertWithDetails(h.UUID("sid"), "testing", "details")
	msg := h.Slack().Channel("test").ExpectMessage("testing")
	msg.AssertColor("#862421")
	msg.AssertActions()
	a.Ack()

	updated := msg.ExpectUpdate()
	updated.AssertText("Ack", "testing")
	updated.AssertNotText("details")
	updated.AssertColor("#867321")
	updated.AssertActions()

	a.Escalate()

	updated = msg.ExpectUpdate()
	updated.AssertText("Escalated", "testing")
	updated.AssertNotText("details")
	updated.AssertColor("#862421")
	updated.AssertActions()
	msg.ExpectBroadcastReply("testing")

	a.Close()

	updated = msg.ExpectUpdate()
	updated.AssertText("Closed", "testing")
	updated.AssertNotText("details")
	updated.AssertColor("#218626")

	updated.AssertActions() // no actions
}
