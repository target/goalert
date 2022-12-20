package smoke

import (
	"testing"

	"github.com/target/goalert/test/smoke/harness"
)

// TestMessageBundle_Voice checks that SMS status updates and alert notifications are bundled when General.MessageBundles is enabled.
func TestMessageBundle_Voice(t *testing.T) {
	t.Parallel()

	sql := `
		insert into users (id, role, name, email) 
		values 
			({{uuid "user"}}, 'user', 'bob', 'joe');
		insert into user_contact_methods (id, user_id, name, type, value) 
		values
			({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'VOICE', {{phone "1"}});
		insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
		values
			({{uuid "user"}}, {{uuid "cm1"}}, 0);
		update users set alert_status_log_contact_method_id = {{uuid "cm1"}}
		where id = {{uuid "user"}};

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
			({{uuid "sid"}}, {{uuid "eid"}}, 'My Service');
`
	h := harness.NewHarness(t, sql, "message-bundles")
	defer h.Close()

	h.SetConfigValue("General.DisableMessageBundles", "false")

	h.CreateAlert(h.UUID("sid"), "test1")
	h.CreateAlert(h.UUID("sid"), "test2")
	h.CreateAlert(h.UUID("sid"), "test3")
	h.CreateAlert(h.UUID("sid"), "test4")

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))

	d1.ExpectCall().Answer().ExpectSay("My Service", "4 unacknowledged").
		Press("4").
		ExpectSay("Acknowledged all").
		Hangup()

	h.GraphQLQuery2(`mutation{ updateAlerts(input: {alertIDs: [1,2,3,4], newStatus: StatusClosed}){id} }`)
}
