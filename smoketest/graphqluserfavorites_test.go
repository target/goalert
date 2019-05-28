package smoketest

import (
	"encoding/json"
	"fmt"
	"github.com/target/goalert/smoketest/harness"
	"testing"
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

	doQL := func(t *testing.T, query string, res interface{}) {
		g := h.GraphQLQueryT(t, query, "/v1/graphql")
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

	doQL(t, fmt.Sprintf(`
		mutation {
			setUserFavorite (input: {
				target_type: service ,target_id: "%s"
			}) {
				target_id
			}
		}
	`, h.UUID("sid1")), nil)

	var s struct {
		Service struct {
			IsUserFav bool `json:"is_user_favorite"`
		}
	}

	doQL(t, fmt.Sprintf(`
		query {
			service(id: "%s") { 
				is_user_favorite 
			}	
		}
	`, h.UUID("sid1")), &s)

	if s.Service.IsUserFav != true {
		t.Fatalf("ERROR: ServiceID %s IsUserFavorite=%t; want true", h.UUID("sid1"), s.Service.IsUserFav)
	}

	doQL(t, fmt.Sprintf(`
		query {
			service(id: "%s") { 
				is_user_favorite 
			}	
		}
	`, h.UUID("sid2")), &s)

	if s.Service.IsUserFav != false {
		t.Fatalf("ERROR: ServiceID %s IsUserFavorite=%t; want false", h.UUID("sid2"), s.Service.IsUserFav)
	}

	// Again Setting as user-favorite should result in no change
	doQL(t, fmt.Sprintf(`
		mutation {
			setUserFavorite (input: {
				target_type: service ,target_id: "%s"
			}) {
				target_id
			}
		}
	`, h.UUID("sid2")), nil)

	doQL(t, fmt.Sprintf(`
		mutation {
			unsetUserFavorite (input: {
				target_type: service ,target_id: "%s"
			}) {
				target_id
			}
		}
	`, h.UUID("sid2")), nil)

	doQL(t, fmt.Sprintf(`
		query {
			service(id: "%s") { 
				is_user_favorite 
			}	
		}
	`, h.UUID("sid2")), &s)

	if s.Service.IsUserFav != false {
		t.Fatalf("ERROR: ServiceID %s IsUserFavorite=%t; want false", h.UUID("sid2"), s.Service.IsUserFav)
	}

}
