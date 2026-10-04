package smoke

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

// TestGraphQLRotationScheduleOnly tests that the Only filter works correctly
// with rotation and schedule search through the GraphQL API.
func TestGraphQLRotationScheduleOnly(t *testing.T) {
	t.Parallel()

	const sql = `
	insert into rotations (id, name, description, type, start_time, time_zone)
	values
		({{uuid "r1"}}, 'rotation one', 'test', 'daily', now(), 'UTC'),
		({{uuid "r2"}}, 'rotation two', 'test', 'daily', now(), 'UTC');
	insert into schedules (id, name, time_zone)
	values
		({{uuid "s1"}}, 'schedule one', 'UTC'),
		({{uuid "s2"}}, 'schedule two', 'UTC');
`

	h := harness.NewHarness(t, sql, "")
	defer h.Close()

	doQL := func(query string, res interface{}) {
		t.Helper()
		g := h.GraphQLQuery2(query)
		for _, err := range g.Errors {
			t.Error("GraphQL Error:", err.Message)
		}
		if len(g.Errors) > 0 {
			t.Fatal("errors returned from GraphQL")
		}

		if res != nil {
			require.NoError(t, json.Unmarshal(g.Data, &res))
		}
	}

	var rotResp struct {
		Rotations struct {
			Nodes []struct{ ID string }
		}
	}
	doQL(fmt.Sprintf(`
		query {
			rotations(input: {only: ["%s"]}) {
				nodes { id }
			}
		}
	`, h.UUID("r1")), &rotResp)
	require.Len(t, rotResp.Rotations.Nodes, 1)
	require.Equal(t, h.UUID("r1"), rotResp.Rotations.Nodes[0].ID)

	var schedResp struct {
		Schedules struct {
			Nodes []struct{ ID string }
		}
	}
	doQL(fmt.Sprintf(`
		query {
			schedules(input: {only: ["%s", "%s"]}) {
				nodes { id }
			}
		}
	`, h.UUID("s1"), h.UUID("s2")), &schedResp)
	require.Len(t, schedResp.Schedules.Nodes, 2)
}
