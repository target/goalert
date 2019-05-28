package smoketest

import (
	"github.com/target/goalert/smoketest/harness"
	"testing"
	"time"
)

// TestUserNotificationCycles tests that the engine
// generates notifications for a notification policy based on the
// 'started_at' timestamp in the 'user_notification_cycles' table
func TestUserNotificationCycles(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "user"}}, 'bob', 'joe');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes, created_at) 
	values
		({{uuid "user"}}, {{uuid "cm1"}}, 0, now()-'1 hour'::interval),
		({{uuid "user"}}, {{uuid "cm1"}}, 1, now()-'1 hour'::interval),
		({{uuid "user"}}, {{uuid "cm1"}}, 5, now()-'1 hour'::interval);

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

	insert into alerts (service_id, description) 
	values
		({{uuid "sid"}}, 'testing');

	insert into notification_logs (id, alert_id, contact_method_id, process_timestamp, completed)
	values
		({{uuid ""}}, 1, {{uuid "cm1"}}, now() - '119 seconds'::interval, true);

	insert into notification_policy_cycles (id, user_id, alert_id, started_at) 
	values
		({{uuid ""}}, {{uuid "user"}}, 1, now() - '120 seconds'::interval);
`
	h := harness.NewHarness(t, sql, "ev3-remove-status-trigger")
	defer h.Close()

	// 0-minute rule should not fire (already sent)
	tw := h.Twilio()
	d1 := tw.Device(h.Phone("1"))

	d1.ExpectSMS("testing") // 1 minute rule should fire (since we're behind)
	h.Delay(15 * time.Second)
	tw.WaitAndAssert()

	h.FastForward(5 * time.Minute)

	// 5-min rule should now fire
	d1.ExpectSMS("testing")
}
