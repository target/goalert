package smoke

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/test/smoke/harness"
)

// TestAlertCleanup verifies that old alerts are purged from the DB
// when `Maintenance.AlertCleanupDays` is set.
func TestAlertCleanup(t *testing.T) {
	t.Parallel()

	sql := `
	insert into escalation_policies (id, name) 
	values
		({{uuid "eid"}}, 'esc policy');
	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');

	insert into alerts (id, service_id, summary, status, created_at) 
	values
		(1, {{uuid "sid"}}, 'testing', 'closed', now()),
		(2, {{uuid "sid"}}, 'testing', 'closed', now() - '2 days'::interval);
`
	h := harness.NewHarness(t, sql, "site24x7-integration")
	defer h.Close()

	h.Trigger()

	var data struct {
		A *struct{ ID string }
		B *struct{ ID string }
	}
	res := h.GraphQLQuery2("{a:alert(id: 1){id} b:alert(id: 2){id}}")
	assert.Empty(t, res.Errors, "errors")
	err := json.Unmarshal(res.Data, &data)
	assert.NoError(t, err)
	assert.Equal(t, "1", data.A.ID)
	assert.Equal(t, "2", data.B.ID)

	cfg := h.Config()
	cfg.Maintenance.AlertCleanupDays = 1
	h.RestartGoAlertWithConfig(cfg)

	assert.EventuallyWithT(t, func(t *assert.CollectT) {
		res = h.GraphQLQuery2("{a:alert(id: 1){id}}")
		assert.Empty(t, res.Errors, "errors")
		err = json.Unmarshal(res.Data, &data)
		assert.NoError(t, err)
		assert.Equal(t, "1", data.A.ID)

		res = h.GraphQLQuery2("{a:alert(id: 2){id}}")
		assert.Empty(t, res.Errors, "errors")
		err = json.Unmarshal(res.Data, &data)
		assert.NoError(t, err)
		// #2 should have been cleaned up
		assert.Nil(t, data.A)
	}, 15*time.Second, time.Second)
}
