package smoke

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

// multiAckSQL builds an escalation policy with a single multi-ack step that
// both users are on. Each user gets an immediate rule and a 30-minute rule.
const multiAckSQL = `
	insert into users (id, name, email, role)
	values
		({{uuid "u1"}}, 'bob', 'bob@example.com', 'user'),
		({{uuid "u2"}}, 'jane', 'jane@example.com', 'user');

	insert into user_contact_methods (id, user_id, name, type, value)
	values
		({{uuid "c1"}}, {{uuid "u1"}}, 'personal', 'SMS', {{phone "1"}}),
		({{uuid "c2"}}, {{uuid "u2"}}, 'personal', 'SMS', {{phone "2"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes)
	values
		({{uuid "u1"}}, {{uuid "c1"}}, 0),
		({{uuid "u1"}}, {{uuid "c1"}}, 30),
		({{uuid "u2"}}, {{uuid "c2"}}, 0),
		({{uuid "u2"}}, {{uuid "c2"}}, 30);

	insert into escalation_policies (id, name)
	values
		({{uuid "eid"}}, 'esc policy');

	insert into escalation_policy_steps (id, escalation_policy_id, multi_ack)
	values
		({{uuid "esid"}}, {{uuid "eid"}}, true);

	insert into escalation_policy_actions (escalation_policy_step_id, user_id)
	values
		({{uuid "esid"}}, {{uuid "u1"}}),
		({{uuid "esid"}}, {{uuid "u2"}});

	insert into services (id, escalation_policy_id, name)
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');

	insert into alerts (id, service_id, summary, dedup_key)
	values
		(198, {{uuid "sid"}}, 'testing', 'auto:1:multiack');
`

// TestMultiAckOtherUsersStillNotified checks that when one user acknowledges an
// alert on a multi-ack step, everyone else on the step keeps getting notified.
func TestMultiAckOtherUsersStillNotified(t *testing.T) {
	t.Parallel()

	h := harness.NewHarness(t, multiAckSQL, "ep-step-multi-ack")
	defer h.Close()

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))
	d2 := tw.Device(h.Phone("2"))

	d1.ExpectSMS("testing")
	d2.ExpectSMS("testing")

	// bob acknowledges; without multi-ack this would cancel jane's 30m rule.
	d1.SendSMS("ack198")
	d1.ExpectSMS("acknowledged")

	h.FastForward(31 * time.Minute)
	h.Trigger()

	// jane is still notified even though the alert is acknowledged.
	d2.ExpectSMS("testing")
}

// TestMultiAckAckerNotNotified checks that the user who acknowledged does not
// keep getting notified by their own remaining notification rules.
func TestMultiAckAckerNotNotified(t *testing.T) {
	t.Parallel()

	h := harness.NewHarness(t, multiAckSQL, "ep-step-multi-ack")
	defer h.Close()

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))
	d2 := tw.Device(h.Phone("2"))

	d1.ExpectSMS("testing")
	d2.ExpectSMS("testing")

	d1.SendSMS("ack198")
	d1.ExpectSMS("acknowledged")

	// jane never acknowledges, so she still gets her 30m rule; bob must not.
	h.FastForward(31 * time.Minute)
	h.Trigger()

	d2.ExpectSMS("testing")
	tw.WaitAndAssert() // fails if bob got a second notification
}

// TestMultiAckSecondAckRecorded checks that a second user acknowledging a
// multi-ack alert is recorded in the alert log rather than being rejected with
// "already acknowledged".
func TestMultiAckSecondAckRecorded(t *testing.T) {
	t.Parallel()

	h := harness.NewHarness(t, multiAckSQL, "ep-step-multi-ack")
	defer h.Close()

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))
	d2 := tw.Device(h.Phone("2"))

	d1.ExpectSMS("testing")
	d2.ExpectSMS("testing")

	d1.SendSMS("ack198")
	d1.ExpectSMS("acknowledged")

	d2.SendSMS("ack198")
	d2.ExpectSMS("acknowledged")

	h.FastForward(time.Minute)

	resp := h.GraphQLQuery2(`{alert(id: 198) {recentEvents(input:{limit:15}){nodes{message}}}}`)
	var respData struct {
		Alert struct {
			RecentEvents struct {
				Nodes []struct {
					Message string
				}
			}
		}
	}
	require.NoError(t, json.Unmarshal(resp.Data, &respData))

	var acks int
	for _, n := range respData.Alert.RecentEvents.Nodes {
		t.Logf("event: %s", n.Message)
		if strings.Contains(n.Message, "Acknowledged") {
			acks++
		}
		// The second ack must not have been rejected as a duplicate.
		assert.NotContains(t, n.Message, "already")
	}

	// Both bob and jane acknowledged, so both must be recorded.
	assert.Equal(t, 2, acks, "expected an alert log entry for each acknowledgment")
}

// TestMultiAckEscalationAfterAck checks that a cycle created *after* a user
// acknowledged is not suppressed by that user's earlier acknowledgment.
//
// bob acks on step 1, the alert is escalated to step 2 (creating a fresh cycle
// for bob), and then jane acks. bob's step 2 cycle started after his own ack,
// so it must survive and deliver his delayed rule.
func TestMultiAckEscalationAfterAck(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email, role)
	values
		({{uuid "u1"}}, 'bob', 'bob@example.com', 'user'),
		({{uuid "u2"}}, 'jane', 'jane@example.com', 'user');

	insert into user_contact_methods (id, user_id, name, type, value)
	values
		({{uuid "c1"}}, {{uuid "u1"}}, 'personal', 'SMS', {{phone "1"}}),
		({{uuid "c2"}}, {{uuid "u2"}}, 'personal', 'SMS', {{phone "2"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes)
	values
		({{uuid "u1"}}, {{uuid "c1"}}, 0),
		({{uuid "u1"}}, {{uuid "c1"}}, 30),
		({{uuid "u2"}}, {{uuid "c2"}}, 0);

	insert into escalation_policies (id, name)
	values
		({{uuid "eid"}}, 'esc policy');

	insert into escalation_policy_steps (id, escalation_policy_id, delay, multi_ack)
	values
		({{uuid "es1"}}, {{uuid "eid"}}, 60, true),
		({{uuid "es2"}}, {{uuid "eid"}}, 60, true);

	insert into escalation_policy_actions (escalation_policy_step_id, user_id)
	values
		({{uuid "es1"}}, {{uuid "u1"}}),
		({{uuid "es2"}}, {{uuid "u1"}}),
		({{uuid "es2"}}, {{uuid "u2"}});

	insert into services (id, escalation_policy_id, name)
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');

	insert into alerts (id, service_id, summary, dedup_key)
	values
		(198, {{uuid "sid"}}, 'testing', 'auto:1:multiack');
`

	h := harness.NewHarness(t, sql, "ep-step-multi-ack")
	defer h.Close()

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))
	d2 := tw.Device(h.Phone("2"))

	// step 1: only bob.
	d1.ExpectSMS("testing")
	d1.SendSMS("ack198")
	d1.ExpectSMS("acknowledged")

	// Escalate to step 2, which creates fresh cycles for bob and jane.
	h.Escalate(198, 0)
	d1.ExpectSMS("testing")
	d2.ExpectSMS("testing")

	// jane acknowledges, putting the alert back into the acknowledged state.
	// bob's step 2 cycle began after his own earlier ack, so it must survive.
	d2.SendSMS("ack198")
	d2.ExpectSMS("acknowledged")

	h.FastForward(31 * time.Minute)
	h.Trigger()

	d1.ExpectSMS("testing")
}

// TestMultiAckCloseStillCancels checks that closing an alert still cancels
// pending notifications, even on a multi-ack step.
func TestMultiAckCloseStillCancels(t *testing.T) {
	t.Parallel()

	h := harness.NewHarness(t, multiAckSQL, "ep-step-multi-ack")
	defer h.Close()

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))
	d2 := tw.Device(h.Phone("2"))

	d1.ExpectSMS("testing")
	d2.ExpectSMS("testing")

	d1.SendSMS("close198")
	d1.ExpectSMS("closed")

	// The 30-minute rules must not fire for anyone.
	h.FastForward(31 * time.Minute)
	h.Trigger()

	tw.WaitAndAssert()
}
