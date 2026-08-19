package smoke

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/expflag"
	"github.com/target/goalert/test/smoke/harness"
)

const (
	signalTestSQL = `
		insert into escalation_policies (id, name) values
			({{uuid "ep"}}, 'esc policy');
		insert into services (id, name, escalation_policy_id) values
			({{uuid "svc"}}, 'service', {{uuid "ep"}});
	`
	signalColorDanger  = "#862421"
	signalColorGood    = "#218626"
	signalColorWarning = "#867321"
	signalColorDefault = "#439FE0"
)

// TestSendSignal tests the sendSignal mutation with a builtin-slack-channel destination.
func TestSendSignal(t *testing.T) {
	t.Parallel()

	h := harness.NewHarnessWithFlags(t, signalTestSQL, "nc-duplicate-table", expflag.FlagSet{expflag.UnivKeys})
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

	msg := h.Slack().Channel("chan1").ExpectMessage("test-signal-message")
	msg.AssertColor(signalColorDefault)
	msg.AssertActions()
}

func TestSendSignalColors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		signalColor string
		wantColor   string
	}{
		{name: "Danger", signalColor: "danger", wantColor: signalColorDanger},
		{name: "Good", signalColor: "good", wantColor: signalColorGood},
		{name: "Warning", signalColor: "warning", wantColor: signalColorWarning},
		{name: "Hex", signalColor: "#439FE0", wantColor: "#439FE0"},
		{name: "Unsupported", signalColor: "unsupported", wantColor: signalColorDefault},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := harness.NewHarnessWithFlags(t, signalTestSQL, "nc-duplicate-table", expflag.FlagSet{expflag.UnivKeys})
			defer h.Close()

			channel := h.Slack().Channel("chan1")
			message := "test-" + test.name + "-signal"
			resp := h.GraphQLQuery2(fmt.Sprintf(`mutation {
				sendSignal(input: {
					serviceID: "%s",
					dest: {type: "builtin-slack-channel", args: {slack_channel_id: "%s"}},
					params: {message: "%s", color: "%s"}
				})
			}`, h.UUID("svc"), channel.ID(), message, test.signalColor))
			require.Empty(t, resp.Errors)

			msg := channel.ExpectMessage(message)
			msg.AssertColor(test.wantColor)
			msg.AssertActions()
		})
	}
}
