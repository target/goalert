package smoketest

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/smoketest/harness"
)

// TestAlertMetrics verifies that ... TODO
func TestAlertMetrics(t *testing.T) {
	t.Parallel()

	sql := `
	insert into escalation_policies (id, name) values ({{uuid "eid"}}, 'ep');
	
	insert into services (id, name, description, escalation_policy_id) 
	values ({{uuid "sid"}}, 'svc', '', {{uuid "eid"}});
	
	insert into alerts (id, summary, status, created_at, service_id) 
	values
		(1,     '', 'closed', now(), {{uuid "sid"}}),
		(2,     '', 'closed', now(), {{uuid "sid"}}),
		(3,     '', 'closed', now(), {{uuid "sid"}}),
		(4,     '', 'closed', now(), {{uuid "sid"}}),
		(505,   '', 'closed', now(), {{uuid "sid"}}),
		(6,     '', 'closed', now(), {{uuid "sid"}});
	
	insert into alert_logs (alert_id, event, message, timestamp)
	values
		(1,     'closed', '', now() - '5 minutes'::interval),
		(2,     'closed', '', now() - '5 minutes'::interval),
		(505,   'closed', '', now() - '5 minutes'::interval),
		(3,     'closed', '', now() - '5 years'::interval),
		(4,     'closed', '', now() - '5 years'::interval),
		(6,     'closed', '', now() - '5 years'::interval);
	`

	h := harness.NewHarness(t, sql, "add-alert-metrics")
	defer h.Close()

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, h.DBURL)
	assert.NoError(t, err)
	defer conn.Close(ctx)

	var metrics_ids []int

	query := func() {
		rows, err := conn.Query(ctx, "select id from alert_metrics")
		assert.NoError(t, err)

		metrics_ids = metrics_ids[:0]
		for rows.Next() {
			var id int
			err := rows.Scan(&id)
			assert.NoError(t, err)
			metrics_ids = append(metrics_ids, id)
		}
	}

	h.Trigger() // cycle 1: recently closed alerts => 1, 2, 505
	h.Trigger() // cycle 2: no state, fill state with 505, check range 5 thru 505 => 6 (505 already processed)
	h.Trigger() // cycle 3: check range 1 thru 5 => 3, 4 (1, 2 already processed)
	query()
	assert.Len(t, metrics_ids, 6)
	assert.Contains(t, metrics_ids, 1)
	assert.Contains(t, metrics_ids, 2)
	assert.Contains(t, metrics_ids, 3)
	assert.Contains(t, metrics_ids, 4)
	assert.Contains(t, metrics_ids, 505)
	assert.Contains(t, metrics_ids, 6)
}
