package smoke

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/expflag"
	"github.com/target/goalert/test/smoke/harness"
)

// TestSendSignal tests the sendSignal mutation with a builtin-slack-channel destination.
func TestSendSignal(t *testing.T) {
	t.Parallel()

	const sql = `
		insert into escalation_policies (id, name) values
			({{uuid "ep"}}, 'esc policy');
		insert into services (id, name, escalation_policy_id) values
			({{uuid "svc"}}, 'service', {{uuid "ep"}});
	`

	h := harness.NewHarnessWithFlags(t, sql, "nc-duplicate-table", expflag.FlagSet{expflag.UnivKeys})
	defer h.Close()

	chanID := h.Slack().Channel("chan1").ID()

	resp := h.GraphQLQuery2(fmt.Sprintf(`mutation {
		sendSignal(input: {
			serviceID: "%s",
			dest: {type: "builtin-slack-channel", args: {slack_channel_id: "%s"}},
			params: {message: "test-signal-message"}
		})
	}`, h.UUID("svc"), chanID))
	require.Empty(t, resp.Errors)

	h.Slack().Channel("chan1").ExpectMessage("test-signal-message")
}
