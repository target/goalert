package smoketest

import (
	"bytes"
	"github.com/target/goalert/smoketest/harness"
	"net/http"
	"net/url"
	"testing"
)

// TestStatusUpdates checks basic functionality of status updates:
//
// - If alert_status_log_contact_method_id isnull, no notifications are sent
// - When alert_status_log_contact_method_id is set, old notifications are NOT sent
// - Status changes, when/after alert_status_log_contact_method_id is set, are sent.
func TestStatusUpdates(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email, role) 
	values 
		({{uuid "user"}}, 'bob', 'joe@test.com', 'admin');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});

	update users set alert_status_log_contact_method_id = {{uuid "cm1"}}
	where id = {{uuid "user"}};

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

	insert into integration_keys (id, service_id, type, name)
	values
		({{uuid "int1"}}, {{uuid "sid"}}, 'generic', 'test');

	insert into alerts (service_id, source, description) 
	values
		({{uuid "sid"}}, 'manual', 'first alert'),
		({{uuid "sid"}}, 'manual', 'second alert');

`
	h := harness.NewHarness(t, sql, "alert-status-updates")
	defer h.Close()

	doClose := func(summary string) {
		u := h.URL() + "/v1/api/alerts?key=" + h.UUID("int1")
		v := make(url.Values)
		v.Set("summary", summary)
		v.Set("action", "close")
		resp, err := http.Post(u, "application/x-www-form-urlencoded", bytes.NewBufferString(v.Encode()))
		if err != nil {
			t.Fatal("post to generic endpoint failed:", err)
		} else if resp.StatusCode/100 != 2 {
			t.Error("non-2xx response:", resp.Status)
		}
		resp.Body.Close()
	}

	tw := h.Twilio()
	d1 := tw.Device(h.Phone("1"))

	d1.ExpectSMS("first alert")
	d1.ExpectSMS("second alert")
	tw.WaitAndAssert()

	doClose("first alert")
	d1.ExpectSMS("closed")
	tw.WaitAndAssert()

	doClose("second alert")
	d1.ExpectSMS("closed")
}
