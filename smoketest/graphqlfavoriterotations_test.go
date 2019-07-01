package smoketest

import (
	"encoding/json"
	"fmt"
	"github.com/target/goalert/smoketest/harness"
	"testing"
)

func TestGraphQLFavoriteRotations(t *testing.T) {
	t.Parallel()
	sql := `
	insert into users (id, name, email, role) 
	values 
		({{uuid "user"}}, 'bob', 'joe', 'admin');

	insert into rotations (id, name, description, type, start_time, time_zone)
	values
		({{uuid "r1"}}, 'test', 'test', 'daily', now(), 'UTC');
`
	h := harness.NewHarness(t, sql, "add-rotation-favorite")
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
			t.Fatal(err)
		}
	}

	doQL(t, fmt.Sprintf(`
	mutation {
   		setFavorite (input: {
			target: {
				id: "%s", 
				type: rotation
			}, 
			favorite: true})
  		}
	`, h.UUID("r1")), nil)
	var r struct {
		Rotation struct {
			IsUserFav bool `json:"isFavorite"`
		}
	}

	doQL(t, fmt.Sprintf(`
		query {
			rotation(id: "%s") { 
				isFavorite 
			}	
		}
	`, h.UUID("r1")), &r)

	if r.Rotation.IsUserFav != true {
		t.Fatalf("ERROR: RotationID %s isFavorite=%t; want true", h.UUID("r1"), r.Rotation.IsUserFav)
	}

	doQL(t, fmt.Sprintf(`
	mutation {
   		setFavorite (input: {
			target: {
				id: "%s", 
				type: rotation
			}, 
			favorite: false})
  		}
	`, h.UUID("r1")), nil)

	doQL(t, fmt.Sprintf(`
		query {
			rotation(id: "%s") { 
				isFavorite 
			}	
		}
	`, h.UUID("r1")), &r)

	if r.Rotation.IsUserFav != false {
		t.Fatalf("ERROR: rotationID %s IsUserFavorite=%t; want false", h.UUID("r1"), r.Rotation.IsUserFav)
	}
}
