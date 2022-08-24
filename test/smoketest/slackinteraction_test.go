package smoketest

import (
	"testing"

	"github.com/target/goalert/test/smoketest/harness"
)

// TestSlackInteraction checks that interactive slack messages work properly.
func TestSlackInteraction(t *testing.T) {
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

	h.SetConfigValue("Slack.InteractiveMessages", "true")

	a := h.CreateAlertWithDetails(h.UUID("sid"), "testing", "details")

	ch := h.Slack().Channel("test")
	msg := ch.ExpectMessage("testing")
	msg.AssertColor("#862421")
	msg.AssertActions("Acknowledge", "Close")

	h.IgnoreErrorsWith("unknown provider/subject")
	msg.Action("Acknowledge").Click() // expect ephemeral
	ch.ExpectEphemeralMessage("GoAlert", "admin")

	h.LinkSlackUser()
	msg.Action("Acknowledge").Click()

	updated := msg.ExpectUpdate()
	updated.AssertText("Ack", "testing")
	updated.AssertColor("#867321")
	updated.AssertActions("Close")

	a.Escalate()

	updated = msg.ExpectUpdate()
	updated.AssertText("Escalated", "testing")
	updated.AssertColor("#862421")
	updated.AssertActions("Acknowledge", "Close")
	msg.ExpectBroadcastReply("testing")

	msg.Action("Close").Click()

	updated = msg.ExpectUpdate()
	updated.AssertText("Closed", "testing")
	updated.AssertNotText("details")
	updated.AssertColor("#218626")

	updated.AssertActions() // no actions
}
