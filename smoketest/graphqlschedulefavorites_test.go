package smoketest

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

// TestGraphQLScheduleFavorites tests that schedules can be set and unset as user favorites
func TestGraphQLScheduleFavorites(t *testing.T) {
	t.Parallel()

	sql := `
	INSERT INTO users (id, name, email) 
	VALUES 
		({{uuid "user"}}, 'gary', 'stu');
	INSERT INTO schedules (id, name, time_zone, description) 
	VALUES
		({{uuid "schedId"}},'schedule1', 'America/Chicago', 'test description here');
`

	h := harness.NewHarness(t, sql, "add-schedule-favorites")
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

	// test setting as a favorite
	doQL(t, fmt.Sprintf(`
		mutation {
			setFavorite(
			  input: { target:{ 
				id: "%s",
				type: schedule 
			  }, 
				favorite: true 
			  }
			)
		  }
	`, h.UUID("schedId")), nil)

	var s struct {
		Schedule struct {
			IsUserFav bool `json:"isFavorite"`
		}
	}

	doQL(t, fmt.Sprintf(`
		query {
			schedule(id: "%s") { 
				isFavorite 
			}	
		}
	`, h.UUID("schedId")), &s)

	if s.Schedule.IsUserFav != true {
		t.Fatalf("ERROR: ScheduleID %s IsUserFavorite=%t; want true", h.UUID("schedId"), s.Schedule.IsUserFav)
	}

	// test unsetting the favorite
	doQL(t, fmt.Sprintf(`
	mutation {
		setFavorite(
		  input: { target:{ 
			id: "%s",
			type: schedule 
		  }, 
			favorite: false
		  }
		)
	  }
	`, h.UUID("schedId")), nil)

	doQL(t, fmt.Sprintf(`
		query {
			schedule(id: "%s") { 
				isFavorite 
			}	
		}
	`, h.UUID("schedId")), &s)

	if s.Schedule.IsUserFav != false {
		t.Fatalf("ERROR: ScheduleID %s IsUserFavorite=%t; want false", h.UUID("schedId"), s.Schedule.IsUserFav)
	}

}
