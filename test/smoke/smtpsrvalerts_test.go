package smoke

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/test/smoke/harness"
)

// TestMailgunAlerts tests that GoAlert responds and
// processes Mailgun requests appropriately.
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

	v := make(url.Values)
	v.Set("recipient", h.UUID("intkey")+"@"+cfg.Mailgun.EmailDomain)
	v.Set("from", "foo@example.com")
	v.Set("subject", "test alert")
	v.Set("body-plain", "details")

	timestamp := time.Now().Format(time.RFC3339)
	token := "some-token"
	v.Set("timestamp", timestamp)
	v.Set("token", token)

	hm := hmac.New(sha256.New, []byte(cfg.Mailgun.APIKey))
	_, _ = io.WriteString(hm, timestamp)
	_, _ = io.WriteString(hm, token)
	calculatedSignature := hm.Sum(nil)

	v.Set("signature", hex.EncodeToString(calculatedSignature))

	resp, err := http.PostForm(h.URL()+"/api/v2/mailgun/incoming", v)
	assert.Nil(t, err)
	if !assert.Equal(t, 200, resp.StatusCode, "create alert (v2 URL)") {
		return
	}

	h.Twilio(t).Device(h.Phone("1")).ExpectSMS("test alert")

	v.Set("subject", "second alert")
	resp, err = http.PostForm(h.URL()+"/v1/webhooks/mailgun", v)
	assert.Nil(t, err)
	if !assert.Equal(t, 200, resp.StatusCode, "create alert (v1 URL)") {
		return
	}

	h.Twilio(t).Device(h.Phone("1")).ExpectSMS("second alert")

	v.Set("recipient", "w"+h.UUID("intkey")+"@"+cfg.Mailgun.EmailDomain)
	resp, err = http.PostForm(h.URL()+"/api/v2/mailgun/incoming", v)
	assert.Nil(t, err)
	if !assert.Equal(t, 406, resp.StatusCode, "reject invalid address with 406 (v2 URL)") {
		return
	}
	// restore
	v.Set("recipient", h.UUID("intkey")+"@"+cfg.Mailgun.EmailDomain)

	v.Set("body-plain", strings.Repeat("too big", 1<<20)) // ~7MiB

	resp, err = http.PostForm(h.URL()+"/api/v2/mailgun/incoming", v)
	assert.Nil(t, err)
	if !assert.Equal(t, 406, resp.StatusCode, "reject large bodies with 406 (v2 URL)") {
		return
	}

	v.Set("body-plain", strings.Repeat("too big", 1<<20)) // ~7MiB

	resp, err = http.PostForm(h.URL()+"/v1/webhooks/mailgun", v)
	assert.Nil(t, err)
	if !assert.Equal(t, 406, resp.StatusCode, "reject large bodies with 406 (v1 URL)") {
		return
	}
}
