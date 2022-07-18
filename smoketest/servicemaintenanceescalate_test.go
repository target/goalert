package smoketest

import (
	"fmt"
	"testing"
	"time"

	"github.com/target/goalert/smoketest/harness"
)

// TestServiceMaintenanceEscalate tests creating and attempting
// to escalate an alert that's already in maintenance mode.
//
// GraphQL should return an error that the service is in maintenance mode.
// No notifications should be sent out to either user on steps 1 and 2.
//
// After turning maintenance mode off, the notification for step 1
// should be sent out.
func TestServiceMaintenanceEscalate(t *testing.T) {
	t.Parallel()

	const initSQL = `
	insert into users (id, name, email)
	values
		({{uuid "user1"}}, 'bob', 'bobby@domain.com'),
		({{uuid "user2"}}, 'joe', 'joseph@domain.com');
	
	insert into user_contact_methods (id, user_id, name, type, value)
	values
		({{uuid "cm1"}}, {{uuid "user1"}}, 'personal', 'SMS', {{phone "1"}}),
		({{uuid "cm2"}}, {{uuid "user2"}}, 'personal', 'SMS', {{phone "2"}});
	
	insert into user_notification_rules (user_id, contact_method_id, delay_minutes)
	values
		({{uuid "user1"}}, {{uuid "cm1"}}, 0),
		({{uuid "user2"}}, {{uuid "cm2"}}, 0);
	
	insert into escalation_policies (id, name)
	values ({{uuid "ep"}}, 'esc policy 1');
	
	insert into escalation_policy_steps (id, escalation_policy_id, delay)
	values
		({{uuid "ep_s1"}}, {{uuid "ep"}}, 10),
		({{uuid "ep_s2"}}, {{uuid "ep"}}, 10);
		
	insert into escalation_policy_actions (escalation_policy_step_id, user_id) 
	values
		({{uuid "ep_s1"}}, {{uuid "user1"}}),
		({{uuid "ep_s2"}}, {{uuid "user2"}});
	
	insert into services (id, escalation_policy_id, name, maintenance_expires_at)
	values ({{uuid "sid"}}, {{uuid "ep"}}, 'service', now() + '1 hour'::interval);`

	h := harness.NewHarness(t, initSQL, "add-service-maintenance-expires-at")
	defer h.Close()

	// create alert
	alert := h.CreateAlert(h.UUID("sid"), "testing")

	// wait an engine cycle, don't expect any SMS messages
	h.Trigger()

	// attempt escalating alert via GraphQL, expect error
	escalateAlert := fmt.Sprintf(`
	mutation {
		escalateAlerts (input: [%d]) {
			id
		}
	}
	`, alert.ID())
	DoGQL(t, h, escalateAlert, nil, "escalate alert: in maintenance mode")

	// turn maintenance mode off, expect message to user on step 1
	setMM := fmt.Sprintf(`
		mutation {
			   updateService (input: {
				id: "%s", 
				maintenanceExpiresAt: "%s"
			})
		  }
		`, h.UUID("sid"), time.Now().Add(-1*time.Hour).Format(time.RFC3339))
	DoGQL(t, h, setMM, nil)

	d := h.Twilio(t).Device(h.Phone("1"))
	d.ExpectSMS("testing")

	// don't expect any SMS from device 2, since that escalation wasn't processed
}
