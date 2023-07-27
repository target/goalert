package smoke

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/test/smoke/harness"
)

func TestAlertLog(t *testing.T) {
	t.Parallel()

	type alertLogs struct {
		Alert struct {
			RecentEvents struct {
				Nodes []struct {
					Message string
					State   struct {
						Details string
					}
				}
			}
		}
	}

	type config struct {
		CMDisabled bool
		CMType     string
		NR         bool
		EPStep     bool
		EPStepUser bool
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

	makeCreateAlertMut := func(h *harness.Harness) string {
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

	const alertLogSQLTmpl = `
		insert into users (id, name, email) 
		values ({{uuid "user"}}, 'bob', 'joe');

		insert into user_contact_methods (id, user_id, name, type, value, disabled) 
		values ({{uuid "cm1"}}, {{uuid "user"}}, 'personal', '{{.CMType}}', {{phone "1"}}, {{.CMDisabled}});

		{{- if .NR}}
		insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
		values ({{uuid "user"}}, {{uuid "cm1"}}, 0);
		{{- end}}

		insert into escalation_policies (id, name) 
		values ({{uuid "eid"}}, 'esc policy');

		{{- if .EPStep}}
		insert into escalation_policy_steps (id, escalation_policy_id) 
		values ({{uuid "esid"}}, {{uuid "eid"}});
		{{- end}}

		{{- if and .EPStep .EPStepUser}}
		insert into escalation_policy_actions (escalation_policy_step_id, user_id) 
		values ({{uuid "esid"}}, {{uuid "user"}});
		{{- end}}

		insert into services (id, escalation_policy_id, name) 
		values ({{uuid "sid"}}, {{uuid "eid"}}, 'service');
	`

	check := func(desc string, c config, before func(*testing.T, *harness.Harness), after func(*testing.T, *harness.Harness, alertLogs)) {
		t.Run(desc, func(t *testing.T) {
			// setup sql
			t.Parallel()
			h := harness.NewHarnessWithData(t, alertLogSQLTmpl, c, "add-no-notification-alert-log")
			defer h.Close()

			// create alert
			doQL(t, h, makeCreateAlertMut(h), nil)

			if before != nil {
				before(t, h)
			}
			h.Trigger()

			// get logs
			var l alertLogs
			doQL(t, h, fmt.Sprintf(`
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
		`, 1), &l)

			// make assertions
			after(t, h, l)
		})
	}

	// test successful notification sent
	check("NotificationSentSuccess", config{
		CMDisabled: false,
		CMType:     "SMS",
		NR:         true,
		EPStep:     true,
		EPStepUser: true,
	}, nil, func(t *testing.T, h *harness.Harness, l alertLogs) {
		msg := l.Alert.RecentEvents.Nodes[0].Message
		assert.Contains(t, msg, "Notification sent")
		h.Twilio(t).Device(h.Phone("1")).ExpectSMS("Alert #1: foo")
	})

	// test disabled contact method
	check("DisabledContactMethod", config{
		CMDisabled: true,
		CMType:     "SMS",
		NR:         true,
		EPStep:     true,
		EPStepUser: true,
	}, nil, func(t *testing.T, h *harness.Harness, l alertLogs) {
		details := l.Alert.RecentEvents.Nodes[0].State.Details
		assert.Contains(t, details, "contact method disabled")
	})

	// test SMS failure
	check("SMSFailure", config{
		CMDisabled: false,
		CMType:     "SMS",
		NR:         true,
		EPStep:     true,
		EPStepUser: true,
	}, func(t *testing.T, h *harness.Harness) {
		h.Twilio(t).Device(h.Phone("1")).RejectSMS("Alert #1: foo")
	}, func(t *testing.T, h *harness.Harness, l alertLogs) {
		msg := l.Alert.RecentEvents.Nodes[0].Message
		details := l.Alert.RecentEvents.Nodes[0].State.Details
		assert.Contains(t, msg, "Notification sent")
		assert.Contains(t, details, "failed")
	})

	// test VOICE failure
	check("VOICEFailure", config{
		CMDisabled: false,
		CMType:     "VOICE",
		NR:         true,
		EPStep:     true,
		EPStepUser: true,
	}, func(t *testing.T, h *harness.Harness) {
		h.Twilio(t).Device(h.Phone("1")).RejectVoice("foo")
	}, func(t *testing.T, h *harness.Harness, l alertLogs) {
		msg := l.Alert.RecentEvents.Nodes[0].Message
		details := l.Alert.RecentEvents.Nodes[0].State.Details
		assert.Contains(t, msg, "Notification sent")
		assert.Contains(t, details, "failed")
	})

	// test no immediate notification rule
	check("NoImmediateNR", config{
		CMDisabled: false,
		CMType:     "SMS",
		NR:         false,
		EPStep:     true,
		EPStepUser: true,
	}, nil, func(t *testing.T, h *harness.Harness, l alertLogs) {
		msg := l.Alert.RecentEvents.Nodes[0].Message
		assert.Contains(t, msg, "no immediate rule")
	})

	// test no on-call users
	check("NoOnCallUsers", config{
		CMDisabled: false,
		CMType:     "SMS",
		NR:         true,
		EPStep:     true,
		EPStepUser: false,
	}, nil, func(t *testing.T, h *harness.Harness, l alertLogs) {
		details := l.Alert.RecentEvents.Nodes[0].State.Details
		assert.Contains(t, details, "No one was on-call")
	})

	// test no steps on an escalation policy
	check("NoEPSteps", config{
		CMDisabled: false,
		CMType:     "SMS",
		NR:         true,
		EPStep:     false,
		EPStepUser: false,
	}, nil, func(t *testing.T, h *harness.Harness, l alertLogs) {
		details := l.Alert.RecentEvents.Nodes[0].State.Details
		assert.Contains(t, details, "No escalation policy steps")
	})
}
