package smoke

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

// TestTwilioSMSAck checks that an SMS ack message is processed.
func TestTwilioSMSAck(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email, role) 
	values 
		({{uuid "user"}}, 'bob', 'joe', 'user');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "user"}}, {{uuid "cm1"}}, 0),
		({{uuid "user"}}, {{uuid "cm1"}}, 30);

	insert into escalation_policies (id, name) 
	values
		({{uuid "eid"}}, 'esc policy');
	insert into escalation_policy_steps (id, escalation_policy_id) 
	values
		({{uuid "esid"}}, {{uuid "eid"}});
	insert into escalation_policy_actions (escalation_policy_step_id, user_id) 
	values 
		({{uuid "esid"}}, {{uuid "user"}});

	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');

	insert into alerts (id, service_id, description) 
	values
		(198, {{uuid "sid"}}, 'testing');

`
	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))

	d1.ExpectSMS("testing").
		ThenReply("ack198").
		ThenExpect("acknowledged")

	h.FastForward(time.Hour)

	resp := h.GraphQLQuery2(`{alert(id: 198) {recentEvents{nodes{message, state{ details, status, formattedSrcValue}}}}}`)
	var respData struct {
		Alert struct {
			RecentEvents struct {
				Nodes []struct {
					Message string
					State   struct {
						Details           string
						Status            string
						FormattedSrcValue string
					}
				}
			}
		}
	}
	err := json.Unmarshal(resp.Data, &respData)
	require.NoError(t, err)
	msgs := respData.Alert.RecentEvents.Nodes
	require.Len(t, msgs, 3)

	t.Logf("msgs: %+v", msgs)

	// note: log is in reverse order
	assert.Contains(t, msgs[0].Message, "Acknowledged by bob")
	assert.Contains(t, msgs[1].Message, "Notification sent to bob")
	assert.Contains(t, msgs[1].State.Details, "delivered")
	assert.Equal(t, "OK", msgs[1].State.Status)
	assert.Contains(t, msgs[2].Message, "Created")
}
