package smoketest

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/target/goalert/smoketest/harness"
)

// TestManyAlerts ensures repeated alert -> msg -> close sequences work properly as long as the delay between
// outlasts max throttle delay.
func TestManyAlerts(t *testing.T) {
	t.Parallel()

	const sql = `
	insert into users (id, name, email, role)
	values
		({{uuid "user"}}, 'bob', 'joe', 'user');

	insert into user_contact_methods (id, user_id, name, type, value)
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes)
	values
		({{uuid "user"}}, {{uuid "cm1"}}, 0);

	insert into escalation_policies (id, name, repeat)
	values
		({{uuid "eid"}}, 'esc policy', 3);

	insert into escalation_policy_steps (id, escalation_policy_id, delay)
	values
		({{uuid "esid"}}, {{uuid "eid"}}, 1);

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

	createAlert := func(summary string) {
		t.Helper()
		u := h.URL() + "/v1/api/alerts?key=" + h.UUID("int_key")
		v := make(url.Values)
		v.Set("summary", summary)
		v.Set("details", "woot")

		resp, err := http.PostForm(u, v)
		if err != nil {
			t.Fatal("post to generic endpoint failed:", err)
		} else if resp.StatusCode/100 != 2 {
			t.Fatal("non-2xx response:", resp.Status)
		}
		resp.Body.Close()
	}

	createAndClose := func(i int) {
		t.Helper()
		summary := fmt.Sprintf("hello-%d-", i)
		createAlert(summary)
		h.FastForward(1 * time.Minute)
		h.FastForward(1 * time.Minute)
		h.FastForward(1 * time.Minute)
		h.Twilio(t).Device(h.Phone("1")).ExpectSMS(summary).ThenReply("c").ThenExpect("closed")
		h.Twilio(t).Device(h.Phone("1")).IgnoreUnexpectedSMS(summary)
		h.FastForward(20 * time.Minute)
	}

	for i := 0; i < 8; i++ {
		createAndClose(i)
	}

}
