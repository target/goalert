package smoketest

import (
	"encoding/json"
	"fmt"
	"github.com/target/goalert/smoketest/harness"
	"testing"
)

// TestGraphQL2Users tests most operations on users API via GraphQL2 endpoint.
func TestGraphQL2Users(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email, role) 
	values 
		({{uuid "user"}}, 'bob', 'joe', 'admin');
`

	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	doQL := func(t *testing.T, query string, res interface{}) {
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

	doQL(t, fmt.Sprintf(`
		mutation {
			addAuthSubject(input: {
				userID: "%s",
				providerID: "%s",
				subjectID: "%s",
			})
		}
	`, h.UUID("user"), "provider1", "subject1"), nil)

	var a struct {
		Nodes struct {
			ProviderID string
			SubjectID  string
			UserID     string
		}
	}

	doQL(t, fmt.Sprintf(`
		query {
			authSubjectsForProvider(providerID: "%s") { 
				nodes {
					providerID
					subjectID
					userID
				}
			}	
		}
	`, "provider1"), &a)

	if a.Nodes.UserID == h.UUID("user") {
		if a.Nodes.SubjectID != "subject1" {
			t.Fatalf("ERROR: retrieved subjectID=%s; want %s", a.Nodes.SubjectID, "subject1")
		}
	}

	doQL(t, fmt.Sprintf(`
		mutation {
			deleteAuthSubject(input: {
				userID: "%s",
				providerID: "%s",
				subjectID: "%s",
			})
		}
	`, h.UUID("user"), "provider1", "subject1"), nil)

	// After deleting provider, no providers should exist
	doQL(t, fmt.Sprintf(`
		query {
			authSubjectsForProvider(providerID: "%s") { 
				nodes {
					providerID
					subjectID
					userID
				}
			}	
		}
	`, "provider1"), &a)

	if len(a.Nodes.ProviderID) != 0 {
		t.Fatalf("ERROR: retrieved Nodes=%s; want nil", a.Nodes.ProviderID)
	}
}
