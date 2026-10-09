package smoke

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

const userCMNRBatchQuery = `
mutation NewCM($input: CreateUserContactMethodInput!) {
	createUserContactMethod(input: $input) {
		id
	}
}

query UsersBatch($ids: [ID!]!) {
	users(input: {only: $ids}) {
		nodes {
			id
			contactMethods {
				id
				name
				private
			}
			notificationRules {
				id
				contactMethodID
			}
		}
	}
}
`

// TestGraphQLUserContactMethodsBatch checks that fetching contactMethods and
// notificationRules for multiple users in a single GraphQL query returns each
// user's own data, correctly keyed, with no cross-user leakage from the
// batched dataloader lookup.
func TestGraphQLUserContactMethodsBatch(t *testing.T) {
	t.Parallel()

	const sql = `
		insert into users (id, name, email, role)
		values
			({{uuid "user1"}}, 'bob', 'joe', 'user'),
			({{uuid "user2"}}, 'bob2', 'joe2', 'user');
	`

	h := harness.NewHarness(t, sql, "")
	defer h.Close()

	newCM := func(userID, name string, private bool) string {
		t.Helper()
		var resp struct {
			CreateUserContactMethod struct{ ID string }
		}
		g := h.GraphQLQueryUserVarsT(t, userID, userCMNRBatchQuery, "NewCM", json.RawMessage(fmt.Sprintf(
			`{"input":{
				"userID": "%s",
				"name": "%s",
				"private": %t,
				"dest":{
					"type":"builtin-smtp-email",
					"args":{"email_address":"%s@example.com"}
				},
				"newUserNotificationRule": {"delayMinutes": 0}
			}}`, userID, name, private, name)))
		require.Empty(t, g.Errors, "expected no errors")
		require.NoError(t, json.Unmarshal(g.Data, &resp))
		return resp.CreateUserContactMethod.ID
	}

	user1CM := newCM(h.UUID("user1"), "user1-cm", false)
	user2CM := newCM(h.UUID("user2"), "user2-cm", false)
	user2PrivCM := newCM(h.UUID("user2"), "user2-priv-cm", true)

	var batch struct {
		Users struct {
			Nodes []struct {
				ID             string
				ContactMethods []struct {
					ID      string
					Name    string
					Private bool
				}
				NotificationRules []struct {
					ID              string
					ContactMethodID string
				}
			}
		}
	}

	// Fetch both users' contact methods and notification rules in a single
	// query, as user1, to prove the batched loader keys results per-user
	// instead of mixing them, and still applies per-user visibility rules.
	g := h.GraphQLQueryUserVarsT(t, h.UUID("user1"), userCMNRBatchQuery, "UsersBatch",
		json.RawMessage(fmt.Sprintf(`{"ids":["%s","%s"]}`, h.UUID("user1"), h.UUID("user2"))))
	require.Empty(t, g.Errors, "expected no errors")
	require.NoError(t, json.Unmarshal(g.Data, &batch))
	require.Len(t, batch.Users.Nodes, 2)

	for _, u := range batch.Users.Nodes {
		switch u.ID {
		case h.UUID("user1"):
			require.Len(t, u.ContactMethods, 1, "user1 should only see their own contact method")
			require.Equal(t, user1CM, u.ContactMethods[0].ID)
			require.Len(t, u.NotificationRules, 1)
			require.Equal(t, user1CM, u.NotificationRules[0].ContactMethodID)
		case h.UUID("user2"):
			require.Len(t, u.ContactMethods, 1, "user1 should not see user2's private contact method")
			require.Equal(t, user2CM, u.ContactMethods[0].ID)
			require.NotEqual(t, user2PrivCM, u.ContactMethods[0].ID)
			require.Len(t, u.NotificationRules, 2, "notification rules are visible regardless of contact method privacy")
		default:
			t.Fatalf("unexpected user ID %s", u.ID)
		}
	}
}
