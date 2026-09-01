package smoke

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

func TestGenericAPIUpdate(t *testing.T) {
	t.Parallel()

	const sql = `
	insert into escalation_policies (id, name)
	values
		({{uuid "ep1"}}, 'policy1');

	insert into escalation_policy_steps (id, escalation_policy_id)
	values
		({{uuid "eps1"}}, {{uuid "ep1"}});

	insert into users (id, name, email)
	values
		({{uuid "u1"}}, 'bob', 'bob@example.com');

	insert into user_contact_methods (id, user_id, name, type, value)
	values
		({{uuid "cm1"}}, {{uuid "u1"}}, 'personal', 'SMS', {{phone "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes)
	values
		({{uuid "u1"}}, {{uuid "cm1"}}, 0);

	insert into escalation_policy_actions (escalation_policy_step_id, user_id)
	values
		({{uuid "eps1"}}, {{uuid "u1"}});

	insert into services (id, escalation_policy_id, name)
	values
		({{uuid "s1"}}, {{uuid "ep1"}}, 'service1');

	insert into integration_keys (id, type, name, service_id)
	values
		({{uuid "i1"}}, 'generic', 'my key', {{uuid "s1"}});
	`

	h := harness.NewHarness(t, sql, "")
	defer h.Close()

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))

	fire := func(summary, dedup string, meta map[string]string, escalate bool) {
		body := map[string]interface{}{
			"summary":  summary,
			"dedup":    dedup,
			"escalate": escalate,
			"meta":     meta,
		}
		data, err := json.Marshal(body)
		require.NoError(t, err)

		resp, err := http.Post(
			h.URL()+"/api/v2/generic/incoming?token="+h.UUID("i1"),
			"application/json",
			bytes.NewReader(data),
		)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	}

	// Create initial alert — user gets notified
	fire("original summary", "dedup1", map[string]string{"sev": "2"}, false)
	d1.ExpectSMS("original summary")

	// Update summary and metadata on existing alert via same dedup key
	fire("updated summary", "dedup1", map[string]string{"sev": "1"}, false)

	// Verify updated values are stored via GraphQL
	res := h.GraphQLQuery2(`query {
		alerts {
			nodes {
				summary
				meta { key value }
			}
		}
	}`)

	var result struct {
		Alerts struct {
			Nodes []struct {
				Summary string
				Meta    []struct {
					Key   string
					Value string
				}
			}
		}
	}
	require.NoError(t, json.Unmarshal(res.Data, &result))
	require.Len(t, result.Alerts.Nodes, 1, "expected exactly one alert")

	alert := result.Alerts.Nodes[0]
	assert.Equal(t, "updated summary", alert.Summary, "summary should be updated")

	var sev string
	for _, m := range alert.Meta {
		if m.Key == "sev" {
			sev = m.Value
		}
	}
	assert.Equal(t, "1", sev, "metadata sev should be updated to 1")

	// Re-escalate the existing alert — user gets notified again with updated summary
	fire("updated summary", "dedup1", map[string]string{"sev": "1"}, true)
	d1.ExpectSMS("updated summary")
}
