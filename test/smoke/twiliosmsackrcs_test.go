package smoke

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

// TestTwilioSMSAck checks that an RCS ack message is processed.
func TestTwilioSMSAckRCS(t *testing.T) {
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
`
	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	rcsSenderID, msgSvcID := h.TwilioMessagingServiceRCS()

	h.SetConfigValue("Twilio.MessagingServiceSID", msgSvcID)
	h.SetConfigValue("Twilio.RCSSenderID", rcsSenderID)

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))

	a := h.CreateAlert(h.UUID("sid"), "testing")
	d1.ExpectSMS("testing").
		ThenReply("ack" + strconv.Itoa(a.ID())).
		ThenExpect("acknowledged")

	h.FastForward(time.Hour)

	assert.EventuallyWithT(t, func(t *assert.CollectT) {
		resp := h.GraphQLQuery2(fmt.Sprintf(`{alert(id: %d) {recentEvents{nodes{message, state{ details, status, formattedSrcValue}}}}}`, a.ID()))
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
		require.Len(t, msgs, 4)

		// note: log is in reverse order
		assert.Contains(t, msgs[0].Message, "Acknowledged by bob")
		assert.Contains(t, msgs[1].Message, "Notification sent to bob")
		assert.Contains(t, msgs[1].State.Details, "read")
		assert.Contains(t, msgs[2].Message, "Escalated")
		assert.Contains(t, msgs[3].Message, "Created")
	}, 15*time.Second, time.Second)
}
