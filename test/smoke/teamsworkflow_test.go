package smoke

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

type teamsWorkflowMessage struct {
	Type        string
	Attachments []struct {
		ContentType string
		Content     json.RawMessage
	}
}

// TestTeamsWorkflow tests that alert notifications and status updates are
// posted to a Microsoft Teams workflow webhook as Adaptive Cards.
func TestTeamsWorkflow(t *testing.T) {
	t.Parallel()

	ch := make(chan teamsWorkflowMessage, 10)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var msg teamsWorkflowMessage

		data, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err) {
			return
		}

		err = json.Unmarshal(data, &msg)
		if !assert.NoError(t, err) {
			return
		}

		ch <- msg
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	sql := `
	insert into escalation_policies (id, name)
	values
		({{uuid "eid"}}, 'esc policy');
	insert into escalation_policy_steps (id, escalation_policy_id)
	values
		({{uuid "esid"}}, {{uuid "eid"}});

	insert into notification_channels (id, name, dest)
	values
		({{uuid "chan"}}, 'teams chan', '{"Type": "builtin-teams-workflow", "Args": {"teams_webhook_url": "` + ts.URL + `"}}');

	insert into escalation_policy_actions (escalation_policy_step_id, channel_id)
	values
		({{uuid "esid"}}, {{uuid "chan"}});

	insert into services (id, escalation_policy_id, name)
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');
`
	h := harness.NewHarness(t, sql, "schedule-rotation-ep-labels")
	defer h.Close()

	expectCard := func(contains ...string) {
		t.Helper()
		timeout := time.After(15 * time.Second)
		select {
		case msg := <-ch:
			assert.Equal(t, "message", msg.Type)
			require.Len(t, msg.Attachments, 1)
			assert.Equal(t, "application/vnd.microsoft.card.adaptive", msg.Attachments[0].ContentType)
			card := string(msg.Attachments[0].Content)
			for _, s := range contains {
				assert.Contains(t, card, s)
			}
		case <-timeout:
			t.Fatal("timed out waiting for teams workflow message")
		}
	}

	a := h.CreateAlertWithDetails(h.UUID("sid"), "testing summary", "testing details")
	h.Trigger()
	expectCard("testing summary", "testing details", "Attention")

	a.Ack()
	h.Trigger()
	expectCard("testing summary", "Warning")

	a.Close()
	h.Trigger()
	expectCard("testing summary", "Good")
}
