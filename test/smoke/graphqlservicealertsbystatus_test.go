package smoke

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

// TestGraphQLServiceAlertsByStatus tests the alertsByStatus field for a Service.
// This test validates that the field returns correct counts for current alert statuses:
// - unacked: alerts with status 'triggered'
// - acked: alerts with status 'active'
// - closed: alerts with status 'closed'
func TestGraphQLServiceAlertsByStatus(t *testing.T) {
	t.Parallel()

	const sql = `
		insert into escalation_policies (id, name)
		values
			({{uuid "eid"}}, 'esc policy');
		insert into services (id, escalation_policy_id, name)
		values
			({{uuid "sid"}}, {{uuid "eid"}}, 'service');
		insert into alerts (id, service_id, status, summary, dedup_key)
		values
			(1, {{uuid "sid"}}, 'triggered', 'alert 1', 'test:1:foo'),
			(2, {{uuid "sid"}}, 'active', 'alert 2', 'test:1:bar'),
			(3, {{uuid "sid"}}, 'active', 'alert 3', 'test:1:baz'),
			(4, {{uuid "sid"}}, 'closed', 'alert 4', null),
			(5, {{uuid "sid"}}, 'closed', 'alert 5', null),
			(6, {{uuid "sid"}}, 'closed', 'alert 6', null);
	`

	h := harness.NewHarness(t, sql, "om-history-index")
	defer h.Close()

	resp := h.GraphQLQueryT(t, fmt.Sprintf(`{service(id: "%s") {alertsByStatus {acked, unacked, closed}}}`, h.UUID("sid")))
	var respValue struct {
		Service struct {
			AlertsByStatus struct {
				Acked   int
				Unacked int
				Closed  int
			}
		}
	}
	t.Logf("Response: %s", resp.Data)
	err := json.Unmarshal(resp.Data, &respValue)
	require.NoError(t, err, "should return valid JSON")
	require.Equal(t, 1, respValue.Service.AlertsByStatus.Unacked, "should have 1 unacked alerts")
	require.Equal(t, 2, respValue.Service.AlertsByStatus.Acked, "should have 2 acked alerts")
	require.Equal(t, 3, respValue.Service.AlertsByStatus.Closed, "should have 3 closed alert")
}
