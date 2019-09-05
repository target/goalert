package smoketest

import (
	"bytes"
	"github.com/target/goalert/smoketest/harness"
	"net/http"
	"testing"
)

func TestSite24x7(t *testing.T) {
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
		({{uuid "int_key"}}, 'site24x7', 'my key', {{uuid "sid"}});
`
	h := harness.NewHarness(t, sql, "site24x7-integration")
	defer h.Close()

	url := h.URL() + "/api/v2/site24x7/incoming?token=" + h.UUID("int_key")

	resp, err := http.Post(url, "application/json", bytes.NewBufferString(`
	{
		"MONITOR_DASHBOARD_LINK": "https://www.site24x7.com/app/client#/home/monitors/xxxxxxxxxxxxxx/Summary",
		"MONITORTYPE": "URL",
		"STATUS": "DOWN",
		"REASON": "Execute on Down",
		"MONITORNAME": "GoAlert Site24x7 Test",
		"ct": "1564479419939",
		"FAILED_LOCATIONS": "Manchester-UK,Edinburgh-UK,Nottingham-UK",
		"INCIDENT_REASON": "Internal Server Error",
		"OUTAGE_TIME_UNIX_FORMAT": "1564479419939",
		"MONITORURL": "https://example.com/test-page",
		"MONITOR_GROUPNAME": "goalert-site24x7-test, URL-goalert-site24x7-test",
		"INCIDENT_TIME": "July 30, 2019 10:36 AM BST",
		"INCIDENT_TIME_ISO": "2019-07-30T10:36:59+0100",
		"RCA_LINK": "https://www.site24x7.com/rca.do?id=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	}
	`))
	if err != nil {
		t.Fatal("post to site24x7 endpoint failed:", err)
	} else if resp.StatusCode != 200 {
		t.Error("non-200 response:", resp.Status)
	}
	resp.Body.Close()

	h.Twilio().Device(h.Phone("1")).ExpectSMS("Site24x7")
}
