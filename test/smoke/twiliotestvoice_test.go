package smoke

import (
	"fmt"
	"testing"

	"github.com/target/goalert/test/smoke/harness"
)

// TestTwilioVoice checks that a test voice call is processed.
func TestTwilioVoice(t *testing.T) {
	t.Parallel()

	sqlQuery := `
	insert into users (id, name, email, role) 
	values 
		({{uuid "user"}}, 'bob', 'joe', 'user');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
	    ({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'VOICE', {{phone "1"}});
`
	h := harness.NewHarness(t, sqlQuery, "add-verification-code")
	defer h.Close()

	doQL := func(query string) {
		g := h.GraphQLQueryUserT(t, h.UUID("user"), query)
		for _, err := range g.Errors {
			t.Error("GraphQL Error:", err.Message)
		}
		if len(g.Errors) > 0 {
			t.Fatal("errors returned from GraphQL")
		}
	}

	doQL(fmt.Sprintf(`
		mutation {
			testContactMethod(id: "%s")
		}
		`, h.UUID("cm1")))

	h.Twilio(t).Device(h.Phone("1")).ExpectVoice("test")
}
