package smoke

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/test/smoke/harness"
)

// TestContactMethodCleanup verifies that stale contact methods are purged from the DB
// when `Maintenance.ContactMethodCleanupDays` is set.
func TestContactMethodCleanup(t *testing.T) {
	t.Parallel()

	sql := `
		insert into users (id, name, email) 
		values 
			({{uuid "user"}}, 'bob', 'joe');
		insert into user_contact_methods (id, user_id, name, type, value, disabled, disabled_since) 
		values
				({{uuid "cm1"}}, {{uuid "user"}}, 'personal1', 'SMS', {{phone "1"}}, true, now() - '2 days'::interval),
				({{uuid "cm2"}}, {{uuid "user"}}, 'personal2', 'SMS', {{phone "2"}}, false, now());
	`
	h := harness.NewHarness(t, sql, "add-timestamps-to-user-contact-methods")
	defer h.Close()

	h.Trigger()

	type data struct {
		User struct {
			ID             string `json:"id"`
			ContactMethods []struct {
				ID string `json:"id"`
			} `json:"contactMethods"`
		} `json:"user"`
	}

	doQL := func(res *data) {
		g := h.GraphQLQuery2(fmt.Sprintf(`
			query {
				user(id: "%s") {
					id
					contactMethods {
						id
					}
				}
			}
		`, h.UUID("user")))
		for _, err := range g.Errors {
			t.Error("GraphQL Error:", err.Message)
		}
		if len(g.Errors) > 0 {
			t.Fatal("errors returned from GraphQL")
		}
		err := json.Unmarshal(g.Data, &res)
		if err != nil {
			t.Fatal("failed to parse response:", err)
		}
	}

	var d data
	doQL(&d)
	assert.Len(t, d.User.ContactMethods, 2)
	assert.Equal(t, h.UUID("cm1"), d.User.ContactMethods[0].ID)
	assert.Equal(t, h.UUID("cm2"), d.User.ContactMethods[1].ID)

	h.SetConfigValue("Maintenance.ContactMethodCleanupDays", "1")

	h.Trigger()

	doQL(&d)
	assert.Len(t, d.User.ContactMethods, 1)
	assert.Equal(t, h.UUID("cm2"), d.User.ContactMethods[0].ID)
}
