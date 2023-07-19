package smoke

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/test/smoke/harness"
)

// TestSMTPAlerts tests that GoAlert responds and
// processes incoming email messages appropriately.
func TestSMTPAlerts(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "user"}}, 'bob', 'joe');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "user"}}, {{uuid "cm1"}}, 0);

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

	insert into integration_keys (id, type, service_id, name) 
	values
		({{uuid "intkey"}}, 'email', {{uuid "sid"}}, 'intkey');
`
	h := harness.NewHarness(t, sql, "trigger-config-sync")
	defer h.Close()

	cfg := h.Config()

	c, err := smtp.Dial(h.SMTPIngressAddr())
	assert.NoError(t, err)
	defer c.Close()

	// create an alert from email
	rcpt := h.UUID("intkey") + "@" + cfg.EmailIngressDomain()
	from := "foo@example.com"
	subj := "test alert"
	date := time.Now().Format(time.RFC3339)
	body := "details"
	msgfmt := "Date: %s\r\nFrom: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s\r\n"
	msg := fmt.Sprintf(msgfmt, date, from, rcpt, subj, body)
	message := strings.NewReader(msg)
	err = c.SendMail(from, []string{rcpt}, message)
	assert.NoError(t, err)

	h.Twilio(t).Device(h.Phone("1")).ExpectSMS("test alert")

	c, err = smtp.Dial(h.SMTPIngressAddr())
	assert.NoError(t, err)
	defer c.Close()

	// validate that invalid addresses are rejected
	rcpt = "w" + h.UUID("intkey") + "@" + cfg.EmailIngressDomain()
	subj = "second alert"
	msg = fmt.Sprintf(msgfmt, date, from, rcpt, subj, body)
	message = strings.NewReader(msg)
	err = c.SendMail(from, []string{rcpt}, message)
	assert.ErrorContains(t, err, "invalid value for 'recipient': bad mailbox name", "reject invalid address")
}
