package smoke

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/test/smoke/harness"
)

// TestTwilioVoiceVerification checks that a verification voice call is processed.
func TestTwilioVoiceVerification(t *testing.T) {
	t.Parallel()

	sqlQuery := `
		insert into users (id, name, email, role) 
		values 
			({{uuid "user"}}, 'bob', 'joe', 'user');
		insert into user_contact_methods (id, user_id, name, type, value, disabled) 
		values
			({{uuid "cm2"}}, {{uuid "user"}}, 'personal', 'VOICE', {{phone "1"}}, true);
		insert into user_notification_rules (id, user_id, delay_minutes, contact_method_id)
		values
			({{uuid "nr2"}}, {{uuid "user"}}, 0, {{uuid "cm2"}});
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

	voiceID := h.UUID("cm2")

	doQL(fmt.Sprintf(`
		mutation {
			sendContactMethodVerification(input:{
				contactMethodID: "%s"
			})
		}
	`, voiceID))
	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))

	call := d1.ExpectVoice("verification")

	codeStr := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, strings.ReplaceAll(call.Body(), "6-digit", ""))

	// Since verification code is said twice during one Twilio message
	assert.Len(t, codeStr, 12)

	code, _ := strconv.Atoi(codeStr[:6])

	doQL(fmt.Sprintf(`
		mutation {
			verifyContactMethod(input:{
				contactMethodID:  "%s",
				code: %d
			})
		}
	`, voiceID, code))

	h.FastForward(time.Minute)

	doQL(fmt.Sprintf(`
		mutation {
			testContactMethod(id: "%s")
		}
	`, voiceID))

	// voice for the given number should be enabled
	d1.ExpectVoice("test")
}
