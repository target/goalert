package smoke

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

// TestGraphQLServiceSearchOnly tests that the Only filter works correctly with service search and pagination through the GraphQL API.
func TestGraphQLServiceSearchOnly(t *testing.T) {
	t.Parallel()

	const sql = `
	insert into escalation_policies (id, name) 
	values
		({{uuid "eid"}}, 'esc policy');
	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid1"}}, {{uuid "eid"}}, 'service one'),
		({{uuid "sid2"}}, {{uuid "eid"}}, 'service two'),
		({{uuid "sid3"}}, {{uuid "eid"}}, 'service three');
`

	h := harness.NewHarness(t, sql, "service-search-only")
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
		t.Log("Response:", string(g.Data))

		if res != nil {
			err := json.Unmarshal(g.Data, &res)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}
		}
	}

	// Query 1: Query with Only filter for 2 services, limit 1
	// should get first service and hasNextPage=true
	query1 := fmt.Sprintf(`
		query {
			services(input: {only: ["%s", "%s"], first: 1}) {
				nodes {
					id
					name
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	`, h.UUID("sid1"), h.UUID("sid2"))

	var resp1 struct {
		Services struct {
			Nodes []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"nodes"`
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
		} `json:"services"`
	}

	doQL(query1, &resp1)

	require.Len(t, resp1.Services.Nodes, 1, "number of returned services")
	require.True(t, resp1.Services.PageInfo.HasNextPage, "expected hasNextPage=true for first query")
	require.NotEmpty(t, resp1.Services.PageInfo.EndCursor, "expected non-empty endCursor for pagination")

	firstServiceID := resp1.Services.Nodes[0].ID
	require.Contains(t, []string{h.UUID("sid1"), h.UUID("sid2")}, firstServiceID, 
		"first service ID should be either sid1 or sid2")

	// Query 2: Use pagination to fetch the second service
	query2 := fmt.Sprintf(`
		query {
			services(input: {only: ["%s", "%s"], first: 1, after: "%s"}) {
				nodes {
					id
					name
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	`, h.UUID("sid1"), h.UUID("sid2"), resp1.Services.PageInfo.EndCursor)

	var resp2 struct {
		Services struct {
			Nodes []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"nodes"`
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
		} `json:"services"`
	}

	doQL(query2, &resp2)

	require.Len(t, resp2.Services.Nodes, 1, "number of returned services in second query")
	require.False(t, resp2.Services.PageInfo.HasNextPage, "expected hasNextPage=false for second query (no more results)")

	secondServiceID := resp2.Services.Nodes[0].ID
	require.Contains(t, []string{h.UUID("sid1"), h.UUID("sid2")}, secondServiceID,
		"second service ID should be either sid1 or sid2")
	require.NotEqual(t, firstServiceID, secondServiceID, "first and second queries should return different services")

	// Query 3: Verify that service 3 is NOT included when using Only filter
	query3 := fmt.Sprintf(`
		query {
			services(input: {only: ["%s", "%s"], first: 10}) {
				nodes {
					id
					name
				}
			}
		}
	`, h.UUID("sid1"), h.UUID("sid2"))

	var resp3 struct {
		Services struct {
			Nodes []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"nodes"`
		} `json:"services"`
	}

	doQL(query3, &resp3)

	require.Len(t, resp3.Services.Nodes, 2, "expected exactly 2 services when querying with Only filter")

	// Verify service 3 is NOT included
	for _, service := range resp3.Services.Nodes {
		require.NotEqual(t, h.UUID("sid3"), service.ID, "service 3 should not be included when using Only filter for services 1 and 2")
	}

	t.Log("All Only filter tests passed successfully!")
}
