package smoke

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

// TestStatusUpdatesNoLog tests status updates continue to work when no logs are present.
func TestStatusUpdatesNoLog(t *testing.T) {
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

	// ensure no processing happens
	err := h.App().Pause(context.Background())
	require.NoError(t, err)
	a.Close()

	_, err = h.App().DB().Exec("delete from alert_logs;")
	require.NoError(t, err)

	err = h.App().Resume(context.Background())
	require.NoError(t, err)

	updated := msg.ExpectUpdate()
	updated.AssertText("Closed", "testing")
	updated.AssertNotText("details")
	updated.AssertColor("#218626")

	updated.AssertActions() // no actions
}
