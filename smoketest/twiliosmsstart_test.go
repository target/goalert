package smoketest

import (
	"fmt"
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

// TestTwilioSMSStart checks that an SMS START message is processed.
func TestTwilioSMSStart(t *testing.T) {
	t.Parallel()

	sql := `
		insert into users (id, name, email) 
		values 
			({{uuid "user"}}, 'bob', 'joe');

		insert into user_contact_methods (id, user_id, name, type, value, disabled) 
		values
			({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}}, false),
			({{uuid "cm2"}}, {{uuid "user"}}, 'personal', 'VOICE', {{phone "1"}}, true);

		insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
		values
			({{uuid "user"}}, {{uuid "cm1"}}, 0),
			({{uuid "user"}}, {{uuid "cm1"}}, 1),
			({{uuid "user"}}, {{uuid "cm2"}}, 1);

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

		insert into alerts (service_id, source, summary, details, dedup_key) 
		values
			({{uuid "sid"}}, 'manual', 'testing', '', 'manual:1:testStart_1');
	`

	h := harness.NewHarness(t, sql, "calendar-subscriptions-per-user") // latest migration
	defer h.Close()

	doQL := func(query string, expectErr bool) {
		g := h.GraphQLQuery2(query)
		if expectErr {
			if len(g.Errors) == 0 {
				t.Fatal("expected error")
			}
			return
		}

		for _, err := range g.Errors {
			t.Error("GraphQL Error:", err.Message)
		}

		if len(g.Errors) > 0 {
			t.Fatal("errors returned from GraphQL")
		}
	}

	d := h.Twilio().Device(h.Phone("1"))
	smsID := h.UUID("cm1")
	voiceID := h.UUID("cm2")

	// setup - disable to get reply text
	d.ExpectSMS("testing").ThenReply("stop")

	// re-enable SMS contact method by text
	d.ExpectSMS("unsubscribed").ThenReply("start")

	// SMS should be enabled
	d.ExpectSMS("re-subscribed")

	// verify by sending test message
	doQL(fmt.Sprintf(`
		mutation {
			testContactMethod(id: "%s")
		}
	`, smsID), false)
	d.ExpectSMS("test")
	h.Twilio().WaitAndAssert()

	// VOICE should still be disabled - expect error
	doQL(fmt.Sprintf(`
		mutation {
			testContactMethod(id: "%s")
		}
	`, voiceID), true)
}
