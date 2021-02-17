package smoketest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/smoketest/harness"
)

type WebhookTestingAlert struct {
	AlertID   int
	AlertType string
	Summary   string
	Details   string
}

func TestWebhookAlert(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var alert WebhookTestingAlert

		jsonbytes, err := ioutil.ReadAll(r.Body)
		assert.Nil(t, err)

		err = json.Unmarshal(jsonbytes, &alert)
		assert.Nil(t, err)

		assert.Equal(t, alert.AlertType, "Alert")
		assert.Equal(t, alert.Summary, "testing summary")
		assert.Equal(t, alert.Details, "testing details")
	}))

	defer ts.Close()

	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "user"}}, 'bob', 'joe');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'WEBHOOK', '` + ts.URL + `');

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

	insert into services (id, escalation_policy_id, name, description) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service', 'testing');

	insert into alerts (service_id, summary, details, status, dedup_key) 
	values
		({{uuid "sid"}}, 'testing summary', 'testing details', 'triggered', 'auto:1:foo');
`

	h := harness.NewHarness(t, sql, "webhook-user-contact-method-type")
	defer h.Close()
}
