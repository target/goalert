package smoketest

import (
	"fmt"
	"github.com/target/goalert/smoketest/harness"
	"testing"
)

// TestTwilioEnableBySMS checks that all contact methods with the same value and of the same user are enabled when the user responds via SMS with the correct code.
func TestTwilioEnableBySMS(t *testing.T) {
	t.Parallel()

	sqlQuery := `
	insert into users (id, name, email) 
	values 
		({{uuid "user"}}, 'bob', 'joe');
	insert into user_contact_methods (id, user_id, name, type, value, disabled) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}}, true),
		({{uuid "cm2"}}, {{uuid "user"}}, 'personal', 'VOICE', {{phone "1"}}, true);
	insert into user_verification_codes (id, user_id, contact_method_value, code, expires_at)
	values 
		({{uuid "id"}}, {{uuid "user"}}, {{phone "1"}}, 123456, now() + '15 minutes'::interval)
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

	cm1 := h.UUID("cm1")
	cm2 := h.UUID("cm2")

	doQL(fmt.Sprintf(`
		mutation {
			verifyContactMethod(input: {
				contact_method_id: "%s",
				verification_code: %d,
			}) {
			  contact_method_ids
			}
		  }
		`, cm1, 123456))

	// All contact methods that have same value and of the same user should be enabled now.
	doQL(fmt.Sprintf(`
				mutation {
					sendContactMethodTest(input:{
						contact_method_id:  "%s",
					}){
						id
					}
				}
				`, cm1))

	d1 := h.Twilio().Device(h.Phone("1"))
	d1.ExpectSMS("test")

	doQL(fmt.Sprintf(`
		mutation {
			sendContactMethodTest(input:{
				contact_method_id:  "%s",
			}){
				id
			}
		}
		`, cm2))

	d1.ExpectVoice("test")
}
