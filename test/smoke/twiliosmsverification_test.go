package smoke

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/target/goalert/test/smoke/harness"
)

// TestTwilioSMSVerification checks that a verification SMS is processed.
func TestTwilioSMSVerification(t *testing.T) {
	t.Parallel()

	sqlQuery := `
		insert into users (id, name, email, role) 
		values 
			({{uuid "user"}}, 'bob', 'joe', 'user');
		insert into user_contact_methods (id, user_id, name, type, value, disabled) 
		values
			({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}}, true);
		insert into user_notification_rules (id, user_id, delay_minutes, contact_method_id)
		values
			({{uuid "nr1"}}, {{uuid "user"}}, 0, {{uuid "cm1"}});
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

	smsID := h.UUID("cm1")

	doQL(fmt.Sprintf(`
		mutation {
			sendContactMethodVerification(input:{
				contactMethodID: "%s"
			})
		}
	`, smsID))
	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))

	msg := d1.ExpectSMS("verification")

	codeStr := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, msg.Body())

	code, _ := strconv.Atoi(codeStr)

	doQL(fmt.Sprintf(`
		mutation {
			verifyContactMethod(input:{
				contactMethodID:  "%s",
				code: %d
			})
		}
	`, smsID, code))

	h.FastForward(time.Minute)

	doQL(fmt.Sprintf(`
		mutation {
			testContactMethod(id: "%s")
		}
	`, smsID))

	// sms for the given number should be enabled
	d1.ExpectSMS("test")
}
