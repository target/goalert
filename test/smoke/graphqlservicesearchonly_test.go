package smoke

import (
	"encoding/json"
	"fmt"
	"testing"

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

	if len(resp1.Services.Nodes) != 1 {
		t.Errorf("Expected 1 service in first query, got %d", len(resp1.Services.Nodes))
	}

	if !resp1.Services.PageInfo.HasNextPage {
		t.Error("Expected hasNextPage=true for first query")
	}

	if resp1.Services.PageInfo.EndCursor == "" {
		t.Error("Expected non-empty endCursor for pagination")
	}

	firstServiceID := resp1.Services.Nodes[0].ID
	if firstServiceID != h.UUID("sid1") && firstServiceID != h.UUID("sid2") {
		t.Errorf("First service ID %s should be either %s or %s", firstServiceID, h.UUID("sid1"), h.UUID("sid2"))
	}

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

	if len(resp2.Services.Nodes) != 1 {
		t.Errorf("Expected 1 service in second query, got %d", len(resp2.Services.Nodes))
	}

	if resp2.Services.PageInfo.HasNextPage {
		t.Error("Expected hasNextPage=false for second query (no more results)")
	}

	secondServiceID := resp2.Services.Nodes[0].ID
	if secondServiceID != h.UUID("sid1") && secondServiceID != h.UUID("sid2") {
		t.Errorf("Second service ID %s should be either %s or %s", secondServiceID, h.UUID("sid1"), h.UUID("sid2"))
	}

	if firstServiceID == secondServiceID {
		t.Error("First and second queries should return different services")
	}

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

	if len(resp3.Services.Nodes) != 2 {
		t.Errorf("Expected exactly 2 services when querying with Only filter, got %d", len(resp3.Services.Nodes))
	}

	for _, service := range resp3.Services.Nodes {
		if service.ID == h.UUID("sid3") {
			t.Error("Service 3 should not be included when using Only filter for services 1 and 2")
		}
	}

	t.Log("All Only filter tests passed successfully!")
}
