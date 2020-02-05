package smoketest

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

// TestGraphQLUserFavorites tests that services can be set and unset as user favorites
func TestGraphQLUserFavorites(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "user"}}, 'bob', 'joe');
	insert into escalation_policies (id, name) 
	values
		({{uuid "eid"}}, 'esc policy');
	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid1"}}, {{uuid "eid"}}, 'service1'),
		({{uuid "sid2"}}, {{uuid "eid"}}, 'service2');
`

	h := harness.NewHarness(t, sql, "UserFavorites")
	defer h.Close()

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
			t.Fatal(err)
		}
	}

	doQL(fmt.Sprintf(`
		mutation{
			setFavorite(input:{
				target:{
					id: "%s"
					type: service
				}
				favorite: true
			})
		}
	`, h.UUID("sid1")), nil)

	var s struct {
		Service struct {
			IsFavoite bool `json:"isFavorite"`
		}
	}

	doQL(fmt.Sprintf(`
		query {
			service(id: "%s") { 
				isFavorite 
			}	
		}
	`, h.UUID("sid1")), &s)

	if s.Service.IsFavoite != true {
		t.Fatalf("ERROR: ServiceID %s IsUserFavorite=%t; want true", h.UUID("sid1"), s.Service.IsFavoite)
	}

	doQL(fmt.Sprintf(`
		query {
			service(id: "%s") { 
				isFavorite 
			}	
		}
	`, h.UUID("sid2")), &s)

	if s.Service.IsFavoite != false {
		t.Fatalf("ERROR: ServiceID %s IsFavoite=%t; want false", h.UUID("sid2"), s.Service.IsFavoite)
	}

	// Again Setting as user-favorite should result in no change
	doQL(fmt.Sprintf(`
		mutation{
			setFavorite(input:{
				target:{
					id: "%s"
					type: service
				}
				favorite: true
			})
		}
	`, h.UUID("sid2")), nil)

	doQL(fmt.Sprintf(`
		mutation{
			setFavorite(input:{
				target:{
					id: "%s"
					type: service
				}
				favorite: false
			})
		}
	`, h.UUID("sid2")), nil)

	doQL(fmt.Sprintf(`
		query {
			service(id: "%s") { 
				isFavorite 
			}	
		}
	`, h.UUID("sid2")), &s)

	if s.Service.IsFavoite != false {
		t.Fatalf("ERROR: ServiceID %s IsUserFavorite=%t; want false", h.UUID("sid2"), s.Service.IsFavoite)
	}

}
