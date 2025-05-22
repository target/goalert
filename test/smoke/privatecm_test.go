package smoke

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

const privCMQuery = `
mutation NewCM($input: CreateUserContactMethodInput!) {
	createUserContactMethod(input: $input) {
		id
	}
}

mutation UpdateCM($input: UpdateUserContactMethodInput!) {
	updateUserContactMethod(input: $input)
}

query GetCM($id: ID!) {
	userContactMethod(id: $id) {
		id
		name
		private
	}
}


query ListUserCM($userID: ID!) {
	user(id: $userID) {
		id
		contactMethods {
			id
			name
			private
		}
		notificationRules {
			id
			contactMethodID
			contactMethod {
				id
			}
		}
	}
}
`

// TestPrivateCM checks that private contact methods are not visible in GraphQL calls from any user that is not the owner.
func TestPrivateCM(t *testing.T) {
	t.Parallel()

	const sql = `
		insert into users (id, name, email)
		values
			({{uuid "user1"}}, 'bob', 'joe'),
			({{uuid "user2"}}, 'bob2', 'joe2');
	`

	h := harness.NewHarness(t, sql, "")
	defer h.Close()

	var newCM struct {
		CreateUserContactMethod struct{ ID string }
	}

	resp := h.GraphQLQueryUserVarsT(t, h.UUID("user1"), privCMQuery, "NewCM",
		json.RawMessage(fmt.Sprintf(
			`{"input":{
				"userID": "%s",
				"name": "cm1",
				"dest":{
					"type":"builtin-smtp-email",
					"args":{"email_address":"foobar@example.com"}
				},
				"newUserNotificationRule": {"delayMinutes": 0}
			}}`, h.UUID("user1"))))
	require.Empty(t, resp.Errors, "expected no errors")
	require.NoError(t, json.Unmarshal(resp.Data, &newCM))
	cmID1 := newCM.CreateUserContactMethod.ID

	resp = h.GraphQLQueryUserVarsT(t, h.UUID("user1"), privCMQuery, "NewCM", json.RawMessage(fmt.Sprintf(
		`{"input":{
			"userID": "%s",
			"name": "cm2",
			"private": true,
			"dest":{
				"type":"builtin-smtp-email",
				"args":{"email_address":"foobar-priv@example.com"}
			},
			"newUserNotificationRule": {"delayMinutes": 0}
		}}`, h.UUID("user1"))))
	require.Empty(t, resp.Errors, "expected no errors")
	require.NoError(t, json.Unmarshal(resp.Data, &newCM))
	cmID2 := newCM.CreateUserContactMethod.ID

	userCanSeeCM := func(userID, cmID string, isPrivate, expectAccess bool) {
		t.Helper()

		// check direct access to the contact method
		resp := h.GraphQLQueryUserVarsT(t, userID, privCMQuery, "GetCM", json.RawMessage(fmt.Sprintf(`{"id":"%s"}`, cmID)))
		require.Empty(t, resp.Errors, "expected no errors")
		t.Log("resp", string(resp.Data))
		var cm struct {
			UserContactMethod *struct {
				ID      string
				Name    string
				Private bool
			}
		}
		require.NoError(t, json.Unmarshal(resp.Data, &cm))
		if !expectAccess {
			require.Nil(t, cm.UserContactMethod, "expected to not see private contact method")
			return
		}
		require.NotNil(t, cm.UserContactMethod, "expected to see contact method")
		require.Equal(t, cmID, cm.UserContactMethod.ID, "expected to see contact method with ID %s", cmID)
		require.Equal(t, isPrivate, cm.UserContactMethod.Private, "expected to see contact method with private=%t", isPrivate)
	}

	userCanSeeCM(h.UUID("user1"), cmID1, false, true)
	userCanSeeCM(h.UUID("user1"), cmID2, true, true)
	userCanSeeCM(h.UUID("user2"), cmID1, false, true)
	userCanSeeCM(h.UUID("user2"), cmID2, true, false)

	// validate listing
	type listCMResp struct {
		User struct {
			ID             string
			ContactMethods []struct {
				ID      string
				Name    string
				Private bool
			}
			NotificationRules []struct {
				ID              string
				ContactMethodID string
				ContactMethod   *struct {
					ID string
				}
			}
		}
	}

	// as user 1
	var listCM listCMResp
	resp = h.GraphQLQueryUserVarsT(t, h.UUID("user1"), privCMQuery, "ListUserCM", json.RawMessage(fmt.Sprintf(`{"userID":"%s"}`, h.UUID("user1"))))
	require.Empty(t, resp.Errors, "expected no errors")

	require.NoError(t, json.Unmarshal(resp.Data, &listCM))

	require.Len(t, listCM.User.ContactMethods, 2, "expected to see both contact methods")
	sort.Slice(listCM.User.ContactMethods, func(i, j int) bool {
		return listCM.User.ContactMethods[i].Name < listCM.User.ContactMethods[j].Name
	})
	require.Equal(t, cmID1, listCM.User.ContactMethods[0].ID, "expected to see contact method 1")
	require.False(t, listCM.User.ContactMethods[0].Private, "expected to see contact method 1 as not private")
	require.True(t, listCM.User.ContactMethods[1].Private, "expected to see contact method 2 as private")
	require.Len(t, listCM.User.NotificationRules, 2, "expected to see two notification rules")
	for _, rule := range listCM.User.NotificationRules {
		switch rule.ContactMethodID {
		case cmID1:
			require.NotNil(t, rule.ContactMethod, "expected to see contact method 1")
		case cmID2:
			require.NotNil(t, rule.ContactMethod, "expected to see contact method 2")
		default:
			t.Fatalf("unexpected contact method ID %s", rule.ContactMethodID)
		}
	}

	// as user 2
	listCM = listCMResp{}
	resp = h.GraphQLQueryUserVarsT(t, h.UUID("user2"), privCMQuery, "ListUserCM", json.RawMessage(fmt.Sprintf(`{"userID":"%s"}`, h.UUID("user1"))))
	require.Empty(t, resp.Errors, "expected no errors")

	require.NoError(t, json.Unmarshal(resp.Data, &listCM))

	require.Len(t, listCM.User.ContactMethods, 1, "expected to see only contact method 1")
	sort.Slice(listCM.User.ContactMethods, func(i, j int) bool {
		return listCM.User.ContactMethods[i].Name < listCM.User.ContactMethods[j].Name
	})
	require.Equal(t, cmID1, listCM.User.ContactMethods[0].ID, "expected to see contact method 1")
	require.False(t, listCM.User.ContactMethods[0].Private, "expected to see contact method 1 as not private")
	require.Len(t, listCM.User.NotificationRules, 2, "expected to see two notification rules")
	for _, rule := range listCM.User.NotificationRules {
		switch rule.ContactMethodID {
		case cmID1:
			require.NotNil(t, rule.ContactMethod, "expected to see contact method 1")
		case cmID2:
			require.Nil(t, rule.ContactMethod, "expected to not see contact method 2")
		default:
			t.Fatalf("unexpected contact method ID %s", rule.ContactMethodID)
		}
	}
}
