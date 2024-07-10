package smoke

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/expflag"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/test/smoke/harness"
)

// TestSignal tests that signal messages are sent correctly.
func TestSignal(t *testing.T) {
	t.Parallel()

	const sql = `
		insert into users (id, name, email) values
			({{uuid "user"}}, 'bob', 'joe');
		insert into user_contact_methods (id, user_id, name, type, value) values
			({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});
		insert into user_notification_rules (user_id, contact_method_id, delay_minutes) values
			({{uuid "user"}}, {{uuid "cm1"}}, 0);
		insert into escalation_policies (id, name) values
			({{uuid "ep"}}, 'esc policy');
		insert into escalation_policy_steps (id, escalation_policy_id, delay) values
			({{uuid "step"}}, {{uuid "ep"}}, 5);
		insert into escalation_policy_actions (escalation_policy_step_id, user_id) values
			({{uuid "step"}}, {{uuid "user"}});
		insert into services (id, name, escalation_policy_id) values
			({{uuid "svc"}}, 'service', {{uuid "ep"}});
	`

	h := harness.NewHarnessWithFlags(t, sql, "nc-duplicate-table", expflag.FlagSet{expflag.UnivKeys})
	defer h.Close()

	var dest gadb.DestV1
	err := h.App().DB().QueryRowContext(context.Background(), `select dest from user_contact_methods where id = $1`, h.UUID("cm1")).Scan(&dest)
	require.NoError(t, err)

	// validate fields
	assert.Equal(t, "builtin-twilio-sms", dest.Type, "unexpected type")
	assert.Equal(t, h.Phone("1"), dest.Arg("phone_number"), "unexpected arg")

	// create key
	resp := h.GraphQLQuery2(fmt.Sprintf(`mutation{ createIntegrationKey(input: {name: "key", type: universal, serviceID: "%s"}){ id, href } }`, h.UUID("svc")))
	require.Empty(t, resp.Errors)

	var respData struct {
		CreateIntegrationKey struct {
			ID   uuid.UUID
			Href string
		}
	}
	err = json.Unmarshal(resp.Data, &respData)
	require.NoError(t, err)

	var gotTestMessage bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method, "unexpected method")
		assert.Equal(t, "/test-path", r.URL.Path, "unexpected webhook path")
		data, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, "webhook-body-data", string(data), "unexpected webhook body")
		gotTestMessage = true
	}))
	defer srv.Close()

	// configure key
	resp = h.GraphQLQuery2(fmt.Sprintf(`
		mutation{
			updateKeyConfig(input: {
				keyID: "%s", 
				defaultActions: [
					{dest: {type: "builtin-alert"},
						params: {summary: "req.body['summary']"}},
					{dest: {type: "builtin-webhook", args: {webhook_url: "%s"}},
						params: {body: "req.body['webhook-body']"}},
					{dest: {type: "builtin-slack-channel", args: {slack_channel_id: "%s"}},
						params: {message: "req.body['slack-text']"}}
				]
			})
		}`, respData.CreateIntegrationKey.ID, srv.URL+"/test-path", h.Slack().Channel("chan1").ID()))
	require.Empty(t, resp.Errors)

	// generate token
	resp = h.GraphQLQuery2(fmt.Sprintf(`mutation{ generateKeyToken(id: "%s")}`, respData.CreateIntegrationKey.ID))
	require.Empty(t, resp.Errors)
	var gen struct {
		GenerateKeyToken string
	}
	err = json.Unmarshal(resp.Data, &gen)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", respData.CreateIntegrationKey.Href, strings.NewReader(`{"summary": "test-summary", "webhook-body": "webhook-body-data", "slack-text": "slack-text-data"}`))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+gen.GenerateKeyToken)
	req.Header.Set("Content-Type", "application/json")
	r, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, r.StatusCode)

	assert.True(t, gotTestMessage, "expected webhook test message")
	h.Slack().Channel("chan1").ExpectMessage("slack-text-data")
	h.Twilio(t).Device(h.Phone("1")).ExpectSMS("test-summary")
}
