package smoke

import (
	"fmt"
	"testing"

	"github.com/target/goalert/test/smoke/harness"
)

// TestTwilioEnablebyVoice checks that all contact methods with the same value and of the same user are enabled when the user responds via Voice with the correct code.
func TestTwilioEnablebyVoice(t *testing.T) {
	t.Parallel()

	sqlQuery := `
		insert into users (id, name, email, role) 
		values 
			({{uuid "user"}}, 'bob', 'joe', 'user');
		insert into user_contact_methods (id, user_id, name, type, value, disabled) 
		values
			({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}}, true),
			({{uuid "cm2"}}, {{uuid "user"}}, 'personal', 'VOICE', {{phone "1"}}, true);
		insert into user_verification_codes (id, user_id, contact_method_value, code, expires_at)
		values 
			({{uuid "id"}}, {{uuid "user"}}, {{phone "1"}}, 123456, now() + '15 minutes'::interval);
		insert into outgoing_messages (message_type, contact_method_id, last_status, sent_at, user_id, user_verification_code_id)
		values
			('verification_message', {{uuid "cm2"}}, 'delivered', now(), {{uuid "user"}}, {{uuid "id"}});
	`
	h := harness.NewHarness(t, sqlQuery, "add-verification-code")
	defer h.Close()

	doQL := func(query string, expectErr bool) {
		g := h.GraphQLQueryUserT(t, h.UUID("user"), query)
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

	smsID := h.UUID("cm1")
	voiceID := h.UUID("cm2")

	doQL(fmt.Sprintf(`
		mutation {
			verifyContactMethod(input:{
				contactMethodID:  "%s",
				code: %d
			})
		}
	`, voiceID, 123456), false)

	// SMS should still be disabled - expect error
	doQL(fmt.Sprintf(`
		mutation {
			testContactMethod(id: "%s")
		}
	`, smsID), true)

	d1 := h.Twilio(t).Device(h.Phone("1"))

	// Voice should now be enabled
	doQL(fmt.Sprintf(`
		mutation {
			testContactMethod(id: "%s")
		}
	`, voiceID), false)

	d1.ExpectVoice("test")
}
