package smoketest

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

func TestPrometheusAlertManager(t *testing.T) {
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
		({{uuid "int_key"}}, 'prometheusAlertmanager', 'my key', {{uuid "sid"}});
`
	h := harness.NewHarness(t, sql, "prometheus-alertmanager-integration")
	defer h.Close()

	url := h.URL() + "/api/v2/prometheusalertmanager/incoming?token=" + h.UUID("int_key")

	resp, err := http.Post(url, "application/json", bytes.NewBufferString(`
	{
		"status": "firing",
		"receiver": "alert-name-receiver-1",
		"externalURL": "http://my.url",
		"alerts": [
			{
				"status": "firing",
				"labels": {"alertname": "TestAlert"},
				"annotations": {"summary": "My alert summary", "description": "My description"}
			}
		]
	}
	`))
	if err != nil {
		t.Fatal("post to prometheus alertmanager endpoint failed:", err)
	} else if resp.StatusCode != 200 {
		t.Error("non-200 response:", resp.Status)
	}
	resp.Body.Close()

	h.Twilio(t).Device(h.Phone("1")).ExpectSMS("bob")
}
