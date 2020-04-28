package smoketest

import (
	"testing"
	"time"

	"github.com/target/goalert/smoketest/harness"
)

// TestEscalation tests that alerts are escalated automatically, per the delay_minutes setting.
func TestEscalation(t *testing.T) {
	t.Parallel()

	const sql = `
insert into users (id, name, email)
values
	({{uuid "user"}}, 'bob', 'joe'),
	({{uuid "user2"}}, 'bob2', 'joe2');

insert into user_contact_methods (id, user_id, name, type, value)
values
	({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}}),
	({{uuid "cm2"}}, {{uuid "user2"}}, 'personal', 'SMS', {{phone "2"}});

insert into user_notification_rules (user_id, contact_method_id, delay_minutes)
values
	({{uuid "user"}}, {{uuid "cm1"}}, 0),
	({{uuid "user2"}}, {{uuid "cm2"}}, 0);

insert into escalation_policies (id, name)
values
	({{uuid "eid"}}, 'esc policy');

insert into escalation_policy_steps (id, escalation_policy_id, delay)
values
	({{uuid "es1"}}, {{uuid "eid"}}, 30),
	({{uuid "es2"}}, {{uuid "eid"}}, 60);
	
insert into escalation_policy_actions (escalation_policy_step_id, user_id) 
values
	({{uuid "es1"}}, {{uuid "user"}}),
	({{uuid "es2"}}, {{uuid "user2"}});

insert into services (id, escalation_policy_id, name)
values
    ({{uuid "sid"}}, {{uuid "eid"}}, 'service');

insert into alerts (service_id, description)
values
    ({{uuid "sid"}}, 'testing');

`
	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	d := h.Twilio(t).Device(h.Phone("1"))
	d.ExpectSMS("testing")
	h.Twilio(t).WaitAndAssert()

	h.FastForward(30 * time.Minute)
	h.Twilio(t).Device(h.Phone("2")).ExpectSMS("testing")
}
