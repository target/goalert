package smoke

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/target/goalert/test/smoke/harness"
)

// TestStatusUpdatesExpiration checks expiration functionality of status updates:
//
// - status updates should not be sent after 7 days of inactivity
func TestStatusUpdatesExpiration(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email, role) 
	values 
		({{uuid "user"}}, 'bob', 'joe@test.com', 'admin');
	insert into user_contact_methods (id, user_id, name, type, value, pending, enable_status_updates) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}}, false, true);

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

	insert into alerts (service_id, source, summary, dedup_key) 
	values
		({{uuid "sid"}}, 'manual', 'first alert', 'user:1:first'),
		({{uuid "sid"}}, 'manual', 'second alert', 'user:1:second');

`
	h := harness.NewHarness(t, sql, "status-update-expiration")
	defer h.Close()

	doClose := func(dedup string) {
		u := h.URL() + "/v1/api/alerts?key=" + h.UUID("int1")
		v := make(url.Values)
		v.Set("dedup", dedup)
		v.Set("action", "close")
		resp, err := http.Post(u, "application/x-www-form-urlencoded", bytes.NewBufferString(v.Encode()))
		if err != nil {
			t.Fatal("post to generic endpoint failed:", err)
		} else if resp.StatusCode/100 != 2 {
			t.Error("non-2xx response:", resp.Status)
		}
		resp.Body.Close()
	}

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))

	d1.ExpectSMS("first alert")
	d1.ExpectSMS("second alert")

	doClose("first")
	d1.ExpectSMS("closed")

	h.FastForward(7 * 24 * time.Hour)

	doClose("second")
}
