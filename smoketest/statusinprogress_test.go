package smoketest

import (
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

// TestStatusInProgress ensures that sent and in-progress notifications for triggered alerts are honored through the migration.
func TestStatusInProgress(t *testing.T) {
	t.Parallel()
	sql := `
		insert into users (id, name, email) 
		values
			({{uuid "u1"}}, 'bob', 'bob@email.com'),
			({{uuid "u2"}}, 'joe', 'joe@email.com');

		insert into user_contact_methods (id, user_id, name, type, value) 
		values
			({{uuid "c1"}}, {{uuid "u1"}}, 'personal', 'SMS', {{phone "1"}});

		update users
		set alert_status_log_contact_method_id = {{uuid "c1"}}
		where id = {{uuid "u1"}};

		insert into escalation_policies (id, name, repeat) 
		values
			({{uuid "eid"}}, 'esc policy', -1);

		insert into services (id, escalation_policy_id, name) 
		values
			({{uuid "sid"}}, {{uuid "eid"}}, 'service');

		insert into alerts (id, service_id, summary, dedup_key) 
		values
			(1, {{uuid "sid"}}, 'test', 'auto:1:foobar');

		insert into alert_logs (id, alert_id, event, sub_user_id, sub_type, message)
		values
			(100, 1, 'created', null, null,''),
			(101, 1, 'notification_sent', {{uuid "u1"}}, 'user', ''),
			(102, 1, 'acknowledged', {{uuid "u2"}}, 'user', ''),
			(103, 1, 'closed', {{uuid "u2"}}, 'user', '');

		truncate user_last_alert_log;

		insert into user_last_alert_log (user_id, alert_id, log_id, next_log_id)
		values
			({{uuid "u1"}}, 1, 102, 103);
	`

	h := harness.NewHarness(t, sql, "sched-module-v3")
	defer h.Close()

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))

	d1.ExpectSMS("Closed", "joe")
}
