package smoketest

import (
	"fmt"
	"github.com/target/goalert/smoketest/harness"
	"strings"
	"testing"
	"time"
)

// TestTwilioVoiceVerification checks that a verification voice call is processed.
func TestTwilioVoiceVerification(t *testing.T) {
	t.Parallel()

	sqlQuery := `
	insert into users (id, name, email) 
	values 
		({{uuid "user"}}, 'bob', 'joe');
	insert into user_contact_methods (id, user_id, name, type, value, disabled) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}}, true),
		({{uuid "cm2"}}, {{uuid "user"}}, 'personal', 'VOICE', {{phone "1"}}, true);
	insert into user_notification_rules (id, user_id, delay_minutes, contact_method_id)
	values
		({{uuid "nr1"}}, {{uuid "user"}}, 0, {{uuid "cm1"}}),
		({{uuid "nr2"}}, {{uuid "user"}}, 0, {{uuid "cm2"}}),
		({{uuid "nr3"}}, {{uuid "user"}}, 1, {{uuid "cm1"}}),
		({{uuid "nr4"}}, {{uuid "user"}}, 1, {{uuid "cm2"}});
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
`
	h := harness.NewHarness(t, sqlQuery, "add-verification-code")
	defer h.Close()

	doQL := func(query string) {
		g := h.GraphQLQuery(query)
		for _, err := range g.Errors {
			t.Error("GraphQL Error:", err.Message)
		}
		if len(g.Errors) > 0 {
			t.Fatal("errors returned from GraphQL")
		}
	}

	cm2 := h.UUID("cm2")

	doQL(fmt.Sprintf(`
		mutation {
			sendContactMethodVerification(input:{
				contact_method_id:  "%s",
			}){
				id
			}
		}
		`, cm2))
	tw := h.Twilio()
	d1 := tw.Device(h.Phone("1"))

	msg := d1.ExpectVoice("verification")
	tw.WaitAndAssert() // wait for code, and ensure no notifications went out

	code := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, msg.Body())

	doQL(fmt.Sprintf(`
		mutation {
			verifyContactMethod(input:{
				contact_method_id:  "%s",
				verification_code: %s
			}){
				contact_method_ids
			}
		}
		`, cm2, code))

	h.FastForward(time.Minute)

	// both CM's for the given number should be enabled
	d1.ExpectSMS("testing")
	d1.ExpectVoice("testing")
}
