package smoke

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/test/smoke/harness"
)

// TestServiceNotices tests notices are properly returned given the appropriate scenarios:
//   - Unacked alert limit reached
//   - Nearing unacked alert limit
func TestServiceNotices(t *testing.T) {
	t.Parallel()

	const sql = `
		insert into escalation_policies (id, name) 
		values
			({{uuid "eid"}}, 'esc policy');

		insert into services (id, escalation_policy_id, name) 
		values
			({{uuid "sid"}}, {{uuid "eid"}}, 'service');

		insert into alerts (id, service_id, summary, created_at, status, dedup_key) 
		values
			(1, {{uuid "sid"}}, 'abc', now(), 'triggered', 'test:1:abc'),
			(2, {{uuid "sid"}}, 'def', now(), 'triggered', 'test:1:def'),
			(3, {{uuid "sid"}}, 'ghi', now(), 'triggered', 'test:1:ghi'),
			(4, {{uuid "sid"}}, 'jkl', now(), 'triggered', 'test:1:jkl'),
			(5, {{uuid "sid"}}, 'mno', now(), 'triggered', 'test:1:mno');
	`

	h := harness.NewHarness(t, sql, "add-pending-to-contact-methods")
	defer h.Close()

	h.SetSystemLimit("unacked_alerts_per_service", 5)

	doQL := func(query string, res interface{}) {
		g := h.GraphQLQuery2(query)
		for _, err := range g.Errors {
			t.Error("GraphQL Error:", err.Message)
		}
		if len(g.Errors) > 0 {
			t.Fatal("errors returned from GraphQL")
		}
		t.Log("Response:", string(g.Data))
		if res == nil {
			return
		}
		err := json.Unmarshal(g.Data, &res)
		if err != nil {
			t.Fatal("failed to parse response:", err)
		}
	}

	sID := h.UUID("sid")
	var notices struct {
		Service struct {
			Notices []struct {
				Type    string
				Message string
				Details string
			}
		}
	}

	query := `
		query {
			service(id: "%s") {
				notices {
					type
					message
					details
				}
			}
		}
	`
	mutation := func(alertIDs string) string {
		return fmt.Sprintf(`
			mutation {
				updateAlerts(input: {
					alertIDs: %v,
					newStatus: StatusClosed
				}) { 
					id
				}
			}
		`, alertIDs)
	}

	doQL(fmt.Sprintf(query, sID), &notices)
	assert.Len(t, notices.Service.Notices, 1)
	assert.Equal(t, notices.Service.Notices[0].Type, "ERROR")
	assert.Contains(t, notices.Service.Notices[0].Message, "Unacknowledged alert limit reached")

	doQL(mutation("[1]"), nil)
	doQL(fmt.Sprintf(query, sID), &notices)
	assert.Len(t, notices.Service.Notices, 1)
	assert.Equal(t, notices.Service.Notices[0].Type, "WARNING")
	assert.Contains(t, notices.Service.Notices[0].Message, "Near unacknowledged alert limit")

	doQL(mutation("[2, 3, 4, 5]"), nil)
	doQL(fmt.Sprintf(query, sID), &notices)
	assert.Empty(t, notices.Service.Notices)
}
