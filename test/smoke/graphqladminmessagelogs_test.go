package smoke

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/test/smoke/harness"
)

// TestGraphQLAdminMessageLogs tests that logs are properly generated for messages.
func TestGraphQLAdminMessageLogs(t *testing.T) {
	t.Parallel()

	const sql = `
		insert into users (id, name, email) 
		values 
			({{uuid "user"}}, 'bob', 'joe');
		insert into user_contact_methods (id, user_id, name, type, value, disabled) 
		values
			({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}}, true);
		insert into outgoing_messages (id, message_type, created_at, sent_at, contact_method_id, last_status, user_id)
		values
			({{uuid "om1"}}, 'test_notification', '2022-01-01 00:01:00', '2022-01-01 00:01:01', {{uuid "cm1"}}, 'delivered', {{uuid "user"}}),
			({{uuid "om2"}}, 'test_notification', '2022-01-01 00:02:00', '2022-01-01 00:01:02', {{uuid "cm1"}}, 'delivered', {{uuid "user"}}),
			({{uuid "om3"}}, 'test_notification', '2022-01-01 00:03:00', null, {{uuid "cm1"}}, 'failed', {{uuid "user"}}),
			({{uuid "om4"}}, 'test_notification', '2022-01-01 00:05:00', '2022-01-01 00:01:05', {{uuid "cm1"}}, 'delivered', {{uuid "user"}}),
			({{uuid "om5"}}, 'test_notification', '2022-01-01 00:04:00', '2022-01-01 00:01:04', {{uuid "cm1"}}, 'delivered', {{uuid "user"}});
	`

	h := harness.NewHarness(t, sql, "switchover-mk2")
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
			t.Fatal("failed to parse response:", err)
		}
	}

	type messageLogs struct {
		MessageLogs struct {
			Nodes []struct {
				ID string `json:"id"`
			} `json:"nodes"`
		} `json:"messageLogs"`
	}

	var logs messageLogs

	doQL(`query {
		messageLogs(input: {}) {
			nodes {
				id
			}
		}
	}`, &logs)

	// tests that the message logs are returned in the correct order
	// of not sent then most recent to least recent
	assert.Len(t, logs.MessageLogs.Nodes, 5, "messageLogs query")
	assert.Equal(t, h.UUID("om3"), logs.MessageLogs.Nodes[0].ID)
	assert.Equal(t, h.UUID("om4"), logs.MessageLogs.Nodes[1].ID)
	assert.Equal(t, h.UUID("om5"), logs.MessageLogs.Nodes[2].ID)
	assert.Equal(t, h.UUID("om2"), logs.MessageLogs.Nodes[3].ID)
	assert.Equal(t, h.UUID("om1"), logs.MessageLogs.Nodes[4].ID)

}
