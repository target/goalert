package smoketest

import (
	"testing"
	"time"

	"github.com/target/goalert/smoketest/harness"
)

// TestInProgress ensures that sent and in-progress notifications for triggered alerts are honored through the migration.
func TestInProgress(t *testing.T) {
	t.Parallel()
	sql := `
    insert into users (id, name, email) 
    values
        ({{uuid "u1"}}, 'bob', 'joe'),
        ({{uuid "u2"}}, 'ben', 'josh'),
        ({{uuid "u3"}}, 'beth', 'jake');

    insert into user_contact_methods (id, user_id, name, type, value) 
    values
        ({{uuid "c1"}}, {{uuid "u1"}}, 'personal', 'SMS', {{phone "1"}}),
        ({{uuid "c1_2"}}, {{uuid "u1"}}, 'personal', 'VOICE', {{phone "1"}}),
        ({{uuid "c2"}}, {{uuid "u2"}}, 'personal', 'SMS', {{phone "2"}}),
        ({{uuid "c2_2"}}, {{uuid "u2"}}, 'personal', 'VOICE', {{phone "2"}}),    
        ({{uuid "c3"}}, {{uuid "u3"}}, 'personal', 'SMS', {{phone "3"}});

    insert into user_notification_rules (id, user_id, contact_method_id, delay_minutes) 
    values
        ({{uuid ""}},{{uuid "u1"}}, {{uuid "c1"}}, 0),
        ({{uuid ""}},{{uuid "u2"}}, {{uuid "c2"}}, 0),
        ({{uuid ""}},{{uuid "u3"}}, {{uuid "c3"}}, 0),
        ({{uuid ""}},{{uuid "u1"}}, {{uuid "c1"}}, 30),
        ({{uuid ""}},{{uuid "u2"}}, {{uuid "c2"}}, 30),
        ({{uuid ""}},{{uuid "u2"}}, {{uuid "c2_2"}}, 30),
        ({{uuid ""}},{{uuid "u3"}}, {{uuid "c3"}}, 30);

    insert into escalation_policies (id, name, repeat) 
    values 
        ({{uuid "eid"}}, 'esc policy', -1);
    insert into escalation_policy_steps (id, escalation_policy_id, delay) 
    values 
        ({{uuid "esid1"}}, {{uuid "eid"}}, 300),
        ({{uuid "esid2"}}, {{uuid "eid"}}, 300),
        ({{uuid "esid3"}}, {{uuid "eid"}}, 300);

    insert into escalation_policy_actions (escalation_policy_step_id, user_id) 
    values
        ({{uuid "esid1"}}, {{uuid "u1"}}),
        ({{uuid "esid1"}}, {{uuid "u2"}}),
        ({{uuid "esid1"}}, {{uuid "u3"}}),
        ({{uuid "esid2"}}, {{uuid "u2"}}),
        ({{uuid "esid3"}}, {{uuid "u3"}});

    insert into services (id, escalation_policy_id, name) 
    values
        ({{uuid "sid"}}, {{uuid "eid"}}, 'service');

    insert into alerts (service_id, summary) 
    values
        ({{uuid "sid"}}, 'testing1'),
        ({{uuid "sid"}}, 'testing2');
    
    insert into escalation_policy_state (alert_id, escalation_policy_id, escalation_policy_step_id, service_id)
    values
        (1, {{uuid "eid"}}, {{uuid "esid1"}}, {{uuid "sid"}});
    
    insert into notification_policy_cycles (alert_id, user_id, last_tick)
    values
        (1, {{uuid "u1"}}, null),
        (1, {{uuid "u2"}}, now() + '35 minutes'::interval),
        (1, {{uuid "u3"}}, now() + '1 second'::interval);
`
	h := harness.NewHarness(t, sql, "UserFavorites")
	defer h.Close()

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))
	d2 := tw.Device(h.Phone("2"))
	d3 := tw.Device(h.Phone("3"))

	d1.ExpectSMS("testing1")
	d1.ExpectSMS("testing2")

	d2.ExpectSMS("testing2")

	d3.ExpectSMS("testing2")

	tw.WaitAndAssert()

	h.FastForward(30 * time.Minute)

	d1.ExpectSMS("testing1")
	d1.ExpectSMS("testing2")

	d2.ExpectSMS("testing2")
	d2.ExpectVoice("testing2")

	d3.ExpectSMS("testing1")
	d3.ExpectSMS("testing2")

	tw.WaitAndAssert()

	h.Escalate(1, 0)

	d2.ExpectSMS("testing1")
	tw.WaitAndAssert()

	h.FastForward(30 * time.Minute)

	d2.ExpectSMS("testing1")
	d2.ExpectVoice("testing1")
	tw.WaitAndAssert()
}
