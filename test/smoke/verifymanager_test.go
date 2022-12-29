package smoke

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/test/smoke/harness"
)

// TestUserVerificationCleanup tests that verification codes are cleaned up
// when they are expired. Also tests that contact methods that are not verified
// are cleaned up with the verification code.
func TestUserVerificationCleanup(t *testing.T) {
	t.Parallel()

	type cm struct {
		ID       string `json:"id"`
		Disabled bool   `json:"disabled"`
		Pending  bool   `json:"pending"`
	}

	type data struct {
		User struct {
			ID             string `json:"id"`
			ContactMethods []cm   `json:"contactMethods"`
		} `json:"user"`
	}

	doQL := func(t *testing.T, h *harness.Harness, query string, res interface{}) {
		t.Helper()
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
	contactMethods := func(t *testing.T, h *harness.Harness) *data {
		t.Helper()
		var d data
		doQL(t, h, fmt.Sprintf(`
			query {
				user(id: "%s") {
					id
					contactMethods {
						id
						disabled
						pending
					}
				}
			}
		`, h.UUID("user")), &d)
		return &d
	}

	checkCM := func(t *testing.T, h *harness.Harness, cm cm, id string, disabled, pending bool) {
		t.Helper()
		assert.Equal(t, id, cm.ID)
		assert.Equal(t, disabled, cm.Disabled)
		assert.Equal(t, pending, cm.Pending)
	}

	t.Run("existing unverified contact method is not deleted for compat", func(t *testing.T) {
		sql := `
			insert into users (id, name, email) 
			values 
				({{uuid "user"}}, 'bob', 'joe');
			insert into user_contact_methods (id, user_id, name, type, value, disabled) 
			values
					({{uuid "cm1"}}, {{uuid "user"}}, 'personal1', 'SMS', {{phone "1"}}, true);
			insert into user_verification_codes (id, contact_method_id, code, expires_at)
			values
				({{uuid "code"}}, {{uuid "cm1"}}, '1234', now() + '15 minutes'::interval);
	`
		h := harness.NewHarness(t, sql, "switchover-mk2")
		defer h.Close()
		h.Migrate("add-pending-to-contact-methods")
		h.FastForward(20 * time.Minute)
		h.Trigger()

		// verify that a unverified contact method created before the migration is not deleted for compat
		d := contactMethods(t, h)
		assert.Len(t, d.User.ContactMethods, 1)
		checkCM(t, h, d.User.ContactMethods[0], h.UUID("cm1"), true, false)
	})

	t.Run("database trigger on disabled field should set pending", func(t *testing.T) {
		sql := `
			insert into users (id, name, email) 
			values 
				({{uuid "user"}}, 'bob', 'joe');
			insert into user_contact_methods (id, user_id, name, type, value, disabled, pending) 
			values
					({{uuid "cm1"}}, {{uuid "user"}}, 'personal1', 'SMS', {{phone "1"}}, true, true);
	`
		h := harness.NewHarness(t, sql, "add-pending-to-contact-methods")
		defer h.Close()

		permission.SudoContext(context.Background(), func(ctx context.Context) {
			err := h.App().ContactMethodStore.EnableByValue(ctx, "SMS", h.Phone("1"))
			require.NoError(t, err)
		})

		// verify that a unverified contact method created before the migration is not deleted for compat
		d := contactMethods(t, h)
		assert.Len(t, d.User.ContactMethods, 1)
		checkCM(t, h, d.User.ContactMethods[0], h.UUID("cm1"), false, false)
	})

	t.Run("new contact methods are cleaned up properly", func(t *testing.T) {
		sql := `
			insert into users (id, name, email) 
			values 
				({{uuid "user"}}, 'bob', 'joe');
			insert into user_contact_methods (id, user_id, name, type, value, disabled, pending) 
			values
					({{uuid "cm1"}}, {{uuid "user"}}, 'personal1', 'SMS', {{phone "1"}}, true, true),
					({{uuid "cm2"}}, {{uuid "user"}}, 'personal2', 'SMS', {{phone "2"}}, false, false);
			insert into user_verification_codes (id, contact_method_id, code, expires_at)
			values
				({{uuid "code"}}, {{uuid "cm1"}}, '1234', now() + '15 minutes'::interval);
	`
		h := harness.NewHarness(t, sql, "add-pending-to-contact-methods")
		defer h.Close()
		h.FastForward(20 * time.Minute)
		h.Trigger()

		d := contactMethods(t, h)
		// cm1 should be deleted as the verification code is expired
		// cm2 should still exist as it is not disabled and not pending
		assert.Len(t, d.User.ContactMethods, 1)
		checkCM(t, h, d.User.ContactMethods[0], h.UUID("cm2"), false, false)

		// disable and re-verify cm2 and let it expire to make sure it is not deleted as it was previously verified
		permission.SudoContext(context.Background(), func(ctx context.Context) {
			err := h.App().ContactMethodStore.DisableByValue(ctx, "SMS", h.Phone("2"))
			require.NoError(t, err)
		})

		d = contactMethods(t, h)
		// cm2 should be disabled and not pending
		assert.Len(t, d.User.ContactMethods, 1)
		checkCM(t, h, d.User.ContactMethods[0], h.UUID("cm2"), true, false)

		// re-verify cm2 and fast forward past the expiration
		doQL(t, h, fmt.Sprintf(`
			mutation {
				sendContactMethodVerification(input: {
					contactMethodID: "%s"
				})
			}
		`, h.UUID("cm2")), nil)

		h.FastForward(20 * time.Minute)
		h.Trigger()

		d = contactMethods(t, h)
		// cm2 should exist, be disabled and not pending
		assert.Len(t, d.User.ContactMethods, 1)
		checkCM(t, h, d.User.ContactMethods[0], h.UUID("cm2"), true, false)
	})
}
