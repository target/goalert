package smoke

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/test/smoke/harness"
)

// TestAlertAutoClosePolicyUpdated verifies that `policy_updated` alert log entries
// are not considered activity when determining if an alert should be auto-closed.
func TestAlertAutoClosePolicyUpdated(t *testing.T) {
	t.Parallel()

	sql := `
	insert into escalation_policies (id, name)
	values
		({{uuid "eid"}}, 'esc policy');
	insert into services (id, escalation_policy_id, name)
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');

	insert into alerts (id, service_id, summary, status, dedup_key, created_at)
	values
		(1, {{uuid "sid"}}, 'no-activity', 'triggered', 'test:1:foo', now() - '2 days'::interval),
		(2, {{uuid "sid"}}, 'policy-updated', 'triggered', 'test:1:bar', now() - '2 days'::interval),
		(3, {{uuid "sid"}}, 'escalated', 'triggered', 'test:1:baz', now() - '2 days'::interval);

	insert into alert_logs (alert_id, event, message, timestamp)
	values
		(2, 'policy_updated', '', now()),
		(3, 'escalated', '', now());
`
	h := harness.NewHarness(t, sql, "")
	defer h.Close()

	const query = "{a:alert(id: 1){status} b:alert(id: 2){status} c:alert(id: 3){status}}"
	type alertStatus struct{ Status string }
	var data struct{ A, B, C alertStatus }

	cfg := h.Config()
	cfg.Maintenance.AlertAutoCloseDays = 1
	h.RestartGoAlertWithConfig(cfg)

	assert.EventuallyWithT(t, func(t *assert.CollectT) {
		res := h.GraphQLQuery2(query)
		assert.Empty(t, res.Errors, "errors")
		assert.NoError(t, json.Unmarshal(res.Data, &data))
		assert.Equal(t, "StatusClosed", data.A.Status, "alert with no activity")
		assert.Equal(t, "StatusClosed", data.B.Status, "alert with only a recent policy update")
	}, 15*time.Second, time.Second)

	// alert 3 has recent activity, so it should remain open
	res := h.GraphQLQuery2(query)
	assert.Empty(t, res.Errors, "errors")
	assert.NoError(t, json.Unmarshal(res.Data, &data))
	assert.Equal(t, "StatusUnacknowledged", data.C.Status, "alert with recent escalation")
}
