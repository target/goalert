package smoketest

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/smoketest/harness"
)

// DONE !!!
func TestNotificationSentSuccess(t *testing.T) {
	t.Parallel()

	var sql = makeSQL(false, true, true, true)
	h := harness.NewHarness(t, sql, "add-no-notification-alert-log")
	defer h.Close()

	// create alert
	doQL(h, t, makeCreateAlertMut(h), nil)
	h.Trigger()
	logs := getLogs(h, t)

	// most recent entry
	var msg = logs.Alert.RecentEvents.Nodes[0].Message
	assert.Contains(t, msg, "Notification sent")
	h.Twilio(t).Device(h.Phone("1")).ExpectSMS("Alert #1: foo")
}

// DONE !!!
func TestDisabledContactMethod(t *testing.T) {
	t.Parallel()

	var sql = makeSQL(true, true, true, true)
	h := harness.NewHarness(t, sql, "add-no-notification-alert-log")
	defer h.Close()

	// create alert
	doQL(h, t, makeCreateAlertMut(h), nil)
	h.Trigger()
	logs := getLogs(h, t)

	// most recent entry
	var details = logs.Alert.RecentEvents.Nodes[0].State.Details
	assert.Equal(t, "contact method disabled", details)
}

func TestSMSFailure(t *testing.T) {
	t.Parallel()
}

func TestVoiceFailure(t *testing.T) {
	t.Parallel()
}

// DONE !!!
func TestNoImmediateNR(t *testing.T) {
	t.Parallel()

	var sql = makeSQL(false, false, true, true)
	h := harness.NewHarness(t, sql, "add-no-notification-alert-log")
	defer h.Close()

	// create alert
	doQL(h, t, makeCreateAlertMut(h), nil)
	h.Trigger()
	logs := getLogs(h, t)

	// most recent entry
	var msg = logs.Alert.RecentEvents.Nodes[0].Message
	assert.Contains(t, msg, "no immediate rule")
}

// DONE !!!
func TestNoOnCallUsers(t *testing.T) {
	t.Parallel()

	var sql = makeSQL(false, true, true, false)
	h := harness.NewHarness(t, sql, "add-no-notification-alert-log")
	defer h.Close()

	// create alert
	doQL(h, t, makeCreateAlertMut(h), nil)
	h.Trigger()
	logs := getLogs(h, t)

	// most recent entry
	var details = logs.Alert.RecentEvents.Nodes[0].State.Details
	assert.Equal(t, "No one was on-call", details)
}

// DONE !!!
func TestNoEPSteps(t *testing.T) {
	t.Parallel()

	var sql = makeSQL(false, true, false, false)
	h := harness.NewHarness(t, sql, "add-no-notification-alert-log")
	defer h.Close()

	// create alert
	doQL(h, t, makeCreateAlertMut(h), nil)
	h.Trigger()
	logs := getLogs(h, t)

	// most recent entry
	var details = logs.Alert.RecentEvents.Nodes[0].State.Details
	assert.Equal(t, "No escalation policy steps", details)
}

func makeSQL(disabledCM bool, nr bool, epStep bool, epStepUser bool) string {
	// start initial sql with user and cm
	var setupSQL = fmt.Sprintf(`
		insert into users (id, name, email) 
		values ({{uuid "user"}}, 'bob', 'joe');

		insert into user_contact_methods (id, user_id, name, type, value, disabled) 
		values ({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}}, %t);
	`, disabledCM)

	// add notification rule if specified
	if nr {
		fmt.Println("MAKING NR")
		setupSQL = setupSQL + `
			insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
			values ({{uuid "user"}}, {{uuid "cm1"}}, 0);
		`
	}

	// add ep
	setupSQL = setupSQL + `
		insert into escalation_policies (id, name) 
		values ({{uuid "eid"}}, 'esc policy');
	`

	// add ep step if specified
	if epStep {
		setupSQL = setupSQL + `
			insert into escalation_policy_steps (id, escalation_policy_id) 
			values ({{uuid "esid"}}, {{uuid "eid"}});
		`
	}

	// add user to ep step if step and stepUser specified
	if epStep && epStepUser {
		setupSQL = setupSQL + `
			insert into escalation_policy_actions (escalation_policy_step_id, user_id) 
			values ({{uuid "esid"}}, {{uuid "user"}});
		`
	}

	// add service
	setupSQL = setupSQL + `
		insert into services (id, escalation_policy_id, name) 
		values ({{uuid "sid"}}, {{uuid "eid"}}, 'service');
	`

	return setupSQL
}

type logs struct {
	Alert struct {
		RecentEvents struct {
			Nodes []struct {
				Message string `json:"message"`
				State   struct {
					Details string `json:"details"`
				} `json:"state"`
			} `json:"nodes"`
		} `json:"recentEvents"`
	} `json:"alert"`
}

func getLogs(h *harness.Harness, t *testing.T) *logs {
	var result logs

	doQL(h, t, fmt.Sprintf(`
		query {
  			alert(id: %d) {
    			recentEvents(input: { limit: 15 }) {
						nodes {
							message
							state {
								details
							}
						}
    			}
  			}
		}
	`, 1), &result)

	return &result
}

func makeCreateAlertMut(h *harness.Harness) string {
	return fmt.Sprintf(
		`mutation {
			createAlert(input: {
				summary: "foo",
				serviceID: "%s"
			}){ id }
		}`,
		h.UUID("sid"),
	)
}

func doQL(h *harness.Harness, t *testing.T, query string, res interface{}) {
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
