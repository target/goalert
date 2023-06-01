package smoke

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/target/goalert/test/smoke/harness"
)

// TestStatusUpdatesCancel checks that status updates are unsubscribed when updates, or the contact method, are disabled.
func TestStatusUpdatesCancel(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email, role) 
	values 
		({{uuid "user"}}, 'bob', 'joe@test.com', 'admin');
	insert into user_contact_methods (id, user_id, name, type, value, disabled) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}}, false),
		({{uuid "cm2"}}, {{uuid "user"}}, 'personal2', 'SMS', {{phone "2"}}, false);

	update users set alert_status_log_contact_method_id = {{uuid "cm1"}}
	where id = {{uuid "user"}};

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "user"}}, {{uuid "cm1"}}, 0),
		({{uuid "user"}}, {{uuid "cm2"}}, 0);

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

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))
	d2 := tw.Device(h.Phone("2"))

	h.CreateAlert(h.UUID("sid"), "first")
	d1.ExpectSMS("first")
	d2.ExpectSMS("first")
	h.Trigger() // cleanup subscription to cm2, since only cm1 is configured

	h.GraphQLQueryUserT(t, h.UUID("user"), fmt.Sprintf(`
		mutation{
			updateUserContactMethod(input:{id:"%s",enableStatusUpdates: false})
			updateUserContactMethod(input:{id:"%s",enableStatusUpdates: true})
		}`, h.UUID("cm1"), h.UUID("cm2")))

	h.Trigger() // cleanup subscription to cm1, now that cm2 is the only one configured
	doClose("first")
	// no status update as only cm1 was subscribed, and the setting change
	// should have canceled the subscription.

	h.CreateAlert(h.UUID("sid"), "second")
	d1.ExpectSMS("second")
	d2.ExpectSMS("second")

	d2.SendSMS("stop")
	h.Trigger() // cleanup subscription to cm2
	d2.SendSMS("start")

	doClose("second")
	// contact method was canceled, so no status updates should be sent.
}
