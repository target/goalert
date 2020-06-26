package smoketest

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"

	"github.com/target/goalert/smoketest/harness"
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

	resp, err := http.Post(u, "application/x-www-form-urlencoded", bytes.NewBufferString(v.Encode()))
	if err != nil {
		t.Fatal("post to generic endpoint failed:", err)
	} else if resp.StatusCode/100 != 2 {
		t.Error("non-2xx response:", resp.Status)
	}
	resp.Body.Close()

	h.Twilio(t).Device(h.Phone("1")).ExpectSMS("hello")
}
