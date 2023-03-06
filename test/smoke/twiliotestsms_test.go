package smoke

import (
	"fmt"
	"testing"

	"github.com/target/goalert/test/smoke/harness"
)

// TestTwilioSMS checks that a test SMS is processed.
func TestTwilioSMS(t *testing.T) {
	t.Parallel()

	sqlQuery := `
	insert into users (id, name, email, role) 
	values 
		({{uuid "user"}}, 'bob', 'joe', 'user');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
	    ({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});
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
	cm1 := h.UUID("cm1")

	doQL(fmt.Sprintf(`
		mutation {
			testContactMethod(id: "%s")
		}
		`, cm1))

	h.Twilio(t).Device(h.Phone("1")).ExpectSMS("test")
}
