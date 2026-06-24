package smoke

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/test/smoke/harness"
)

// TestAlertAutoClose verifies that inactive alerts are closed
// when `Maintenance.AlertAutoCloseDays` is set.
func TestAlertAutoClose(t *testing.T) {
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
		(1, {{uuid "sid"}}, 'testing1', 'triggered', 'test:1:foo', now() - '2 days'::interval),
		(2, {{uuid "sid"}}, 'testing2', 'triggered', 'test:1:bar', now());
`
	h := harness.NewHarness(t, sql, "site24x7-integration")
	defer h.Close()

	h.Trigger()

	var data struct {
		A *struct {
			ID     string
			Status string
		}
		B *struct {
			ID     string
			Status string
		}
	}
	res := h.GraphQLQuery2("{a:alert(id: 1){id, status} b:alert(id: 2){id, status}}")
	assert.Empty(t, res.Errors, "errors")
	err := json.Unmarshal(res.Data, &data)
	assert.NoError(t, err)
	assert.Equal(t, "1", data.A.ID)
	assert.Equal(t, "StatusUnacknowledged", data.A.Status)
	assert.Equal(t, "2", data.B.ID)
	assert.Equal(t, "StatusUnacknowledged", data.B.Status)

	cfg := h.Config()
	cfg.Maintenance.AlertAutoCloseDays = 1
	h.RestartGoAlertWithConfig(cfg)

	assert.EventuallyWithT(t, func(t *assert.CollectT) {
		res = h.GraphQLQuery2("{a:alert(id: 1){id, status} b:alert(id: 2){id, status}}")
		assert.Empty(t, res.Errors, "errors")
		err = json.Unmarshal(res.Data, &data)
		assert.NoError(t, err)
		assert.Equal(t, "1", data.A.ID)
		assert.Equal(t, "StatusClosed", data.A.Status)
		assert.Equal(t, "2", data.B.ID)
		assert.Equal(t, "StatusUnacknowledged", data.B.Status)
	}, 15*time.Second, time.Second)
}
