package smoke

import (
	"fmt"
	"testing"
	"time"

	"github.com/target/goalert/test/smoke/harness"
)

// TestServiceMaintenanceAlert tests that an alert created for a user before
// maintenance mode starts still follows the normal notification policy flow.
//
// Then creates another alert with maintenance mode still on and expects
// no message to be received.
func TestServiceMaintenanceAlert(t *testing.T) {
	t.Parallel()

	const initSQL = `
	insert into users (id, name, email)
	values ({{uuid "user"}}, 'bob', 'bobby@domain.com');

	insert into user_contact_methods (id, user_id, name, type, value)
	values ({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes)
	values ({{uuid "user"}}, {{uuid "cm1"}}, 0), ({{uuid "user"}}, {{uuid "cm1"}}, 5);

	insert into escalation_policies (id, name)
	values ({{uuid "eid"}}, 'esc policy');

	insert into escalation_policy_steps (id, escalation_policy_id, delay)
	values ({{uuid "es1"}}, {{uuid "eid"}}, 5);

	insert into escalation_policy_actions (escalation_policy_step_id, user_id) 
	values ({{uuid "es1"}}, {{uuid "user"}});

	insert into services (id, escalation_policy_id, name)
	values ({{uuid "sid"}}, {{uuid "eid"}}, 'service');`

	h := harness.NewHarness(t, initSQL, "add-service-maintenance-expires-at")
	defer h.Close()

	// contact method phone number to use
	d := h.Twilio(t).Device(h.Phone("1"))

	// create first alert and expect sms as normal
	h.CreateAlert(h.UUID("sid"), "testing")
	d.ExpectSMS("testing")

	// set to maintenance mode
	setMM := fmt.Sprintf(`
	mutation {
   		updateService (input: {
			id: "%s", 
			maintenanceExpiresAt: "%s"
		})
  	}
	`, h.UUID("sid"), time.Now().Add(2*time.Hour).Format(time.RFC3339))
	h.GraphQLQuery2(setMM)

	// current impl, user should still get notified from 2nd notification rule
	h.FastForward(5 * time.Minute)
	d.ExpectSMS("testing")

	// create alert while in maintenance mode
	h.CreateAlert(h.UUID("sid"), "maint mode")

	// no text received
}
