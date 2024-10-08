package smoke

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/target/goalert/test/smoke/harness"
)

func TestHeartbeat(t *testing.T) {
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
		({{uuid "user"}}, {{uuid "cm1"}}, 0),
		({{uuid "user"}}, {{uuid "cm1"}}, 15);

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

	insert into heartbeat_monitors (id, name, service_id, heartbeat_interval, disable_reason)
	values
		({{uuid "hb_key"}}, 'testone', {{uuid "sid"}}, '60 minutes', null),
		({{uuid "hb_key2"}}, 'testtwo', {{uuid "sid"}}, '75 minutes', 'disabled');
`
	h := harness.NewHarness(t, sql, "hb-monitor-disable")
	defer h.Close()

	heartbeat := func(id string) {
		v := make(url.Values)
		v.Set("integrationKey", h.UUID("int_key"))
		resp, err := http.PostForm(h.URL()+"/v1/api/heartbeat/"+id, v)
		if err != nil {
			t.Fatal("post to generic endpoint failed:", err)
		} else if resp.StatusCode/100 != 2 {
			t.Error("non-2xx response:", resp.Status)
		}
		resp.Body.Close()
	}

	// start both monitors
	heartbeat(h.UUID("hb_key"))
	heartbeat(h.UUID("hb_key2"))
	h.FastForward(60 * time.Minute) // expire heartbeat
	h.Twilio(t).Device(h.Phone("1")).ExpectSMS("heartbeat", "testone")

	heartbeat(h.UUID("hb_key")) // reset first
	h.Trigger()                 // cycle engine (to close/process heartbeat) before fast-forwarding

	h.FastForward(15 * time.Minute) // next notification rule
	h.Trigger()
	// no SMS, healthy
	// Note: the second heartbeat monitor is disabled, so it should not trigger an alert.
}
