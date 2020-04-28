package smoketest

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

func TestGenericAPIDedup(t *testing.T) {
	t.Parallel()

	const sql = `
	insert into users (id, name, email)
	values
		({{uuid "u1"}}, 'bob', 'joe'),
		({{uuid "u2"}}, 'bob2', 'joe2');

	insert into user_contact_methods (id, user_id, name, type, value)
	values
		({{uuid "cm1"}}, {{uuid "u1"}}, 'personal', 'SMS', {{phone "1"}}),
		({{uuid "cm2"}}, {{uuid "u2"}}, 'personal', 'SMS', {{phone "2"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes)
	values
		({{uuid "u1"}}, {{uuid "cm1"}}, 0),
		({{uuid "u2"}}, {{uuid "cm2"}}, 0);

	insert into escalation_policies (id, name)
	values
		({{uuid "e1"}}, 'esc policy1'),
		({{uuid "e2"}}, 'esc policy2');

	insert into escalation_policy_steps (id, escalation_policy_id)
	values
		({{uuid "e1s1"}}, {{uuid "e1"}}),
		({{uuid "e2s1"}}, {{uuid "e2"}});

	insert into escalation_policy_actions (escalation_policy_step_id, user_id)
	values
		({{uuid "e1s1"}}, {{uuid "u1"}}),
		({{uuid "e2s1"}}, {{uuid "u2"}});

	insert into services (id, escalation_policy_id, name)
	values
		({{uuid "s1"}}, {{uuid "e1"}}, 'service1'),
		({{uuid "s2"}}, {{uuid "e2"}}, 'service2');

	insert into integration_keys (id, type, name, service_id)
	values
		({{uuid "i1"}}, 'generic', 'my key', {{uuid "s1"}}),
		({{uuid "i2"}}, 'generic', 'my key', {{uuid "s2"}});

	insert into alerts (source, service_id, description)
	values
		('generic', {{uuid "s1"}}, 'pre-existing');
`
	h := harness.NewHarness(t, sql, "add-generic-integration-key")
	defer h.Close()

	fire := func(key, summary, dedup string) {
		u := h.URL() + "/v1/api/alerts?key=" + key
		v := make(url.Values)
		v.Set("summary", summary)
		if dedup != "" {
			v.Set("dedup", dedup)
		}

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
	d2 := tw.Device(h.Phone("2"))

	d1.ExpectSMS("pre-existing") // already created
	tw.WaitAndAssert()

	fire(h.UUID("i1"), "pre-existing", "") // should get deduped and never notify

	fire(h.UUID("i1"), "hello", "")
	fire(h.UUID("i1"), "hello", "") // 1 alert
	d1.ExpectSMS("hello")
	tw.WaitAndAssert()

	fire(h.UUID("i1"), "goodbye", "")
	fire(h.UUID("i2"), "hello", "")
	d1.ExpectSMS("goodbye")
	d2.ExpectSMS("hello") // ensure 2nd service can get an alert
	tw.WaitAndAssert()

	fire(h.UUID("i1"), "hello", "foo")
	fire(h.UUID("i1"), "hello2", "foo")
	d1.ExpectSMS("hello")

	fire(h.UUID("i2"), "hello", "foo")
	d2.ExpectSMS("hello")
	tw.WaitAndAssert()
}
