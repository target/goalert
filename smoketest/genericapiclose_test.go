package smoketest

import (
	"bytes"
	"github.com/target/goalert/smoketest/harness"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestGenericAPIClose(t *testing.T) {
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
		({{uuid "user"}}, {{uuid "cm1"}}, 1);

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

	fire := func(key, summary, dedup string, close bool) {
		u := h.URL() + "/v1/api/alerts?key=" + key
		v := make(url.Values)
		v.Set("summary", summary)
		if dedup != "" {
			v.Set("dedup", dedup)
		}
		if close {
			v.Set("action", "close")
		}

		resp, err := http.Post(u, "application/x-www-form-urlencoded", bytes.NewBufferString(v.Encode()))
		if err != nil {
			t.Fatal("post to generic endpoint failed:", err)
		} else if resp.StatusCode/100 != 2 {
			t.Error("non-2xx response:", resp.Status)
		}
		resp.Body.Close()
	}

	key := h.UUID("int_key")
	fire(key, "test1", "", false)
	fire(key, "test2", "", false)
	fire(key, "test3", "dedup", false)
	fire(key, "test4", "", true) // should not open one in the first place

	d := h.Twilio().Device(h.Phone("1"))

	d.ExpectSMS("test1")
	d.ExpectSMS("test2")
	d.ExpectSMS("test3")
	h.Twilio().WaitAndAssert()

	fire(key, "test2", "", true)
	fire(key, "test3", "dedup", true)

	h.FastForward(time.Minute)

	d.ExpectSMS("test1")
}
