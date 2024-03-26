package smoke

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

func TestGenericAPI(t *testing.T) {
	t.Parallel()

	const sql = `
	insert into users (id, name, email)
	values
		({{uuid "user"}}, 'bob', 'joe');

	insert into user_contact_methods (id, user_id, name, type, value)
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes)
	values
		({{uuid "user"}}, {{uuid "cm1"}}, 0);

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

	insert into integration_keys (id, type, name, service_id)
	values
		({{uuid "int_key"}}, 'generic', 'my key', {{uuid "sid"}});
`
	h := harness.NewHarness(t, sql, "add-generic-integration-key")
	defer h.Close()

	u := h.URL() + "/v1/api/alerts?key=" + h.UUID("int_key")
	v := make(url.Values)
	v.Set("summary", "hello")
	v.Set("details", "woot")
	v.Set("meta", "{\"foo\": \"bar\", \"key\": \"value\"}")

	resp, err := http.Post(u, "application/x-www-form-urlencoded", bytes.NewBufferString(v.Encode()))
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode, "http status code")
	resp.Body.Close()

	h.Twilio(t).Device(h.Phone("1")).ExpectSMS("hello")

	resp, err = http.Post(u, "application/json", strings.NewReader(`{"summary": "json"}`))
	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode, "http status code")
	resp.Body.Close()

	h.Twilio(t).Device(h.Phone("1")).ExpectSMS("json")

	req, err := http.NewRequest("POST", u, strings.NewReader(`{"summary": "json"}`))
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode, "http status code")

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, `{"AlertID":2,"ServiceID":"`+h.UUID("sid")+`","IsNew":false}`, string(data), "json response (duplicate alert)")
}
