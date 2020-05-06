package smoketest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/limit"
	"github.com/target/goalert/smoketest/harness"
)

// TestSystemLimits tests that limits are enforced if configured.
func TestSystemLimits(t *testing.T) {
	t.Parallel()

	const sql = `
		insert into users (id, name)
		values
			({{uuid "cm_user"}}, 'CM User'),
			({{uuid "nr_user"}}, 'NR User'),
			('50322144-1e88-43dc-b638-b16a5be7bad6', 'User 1'),
			('dfcc0684-f045-4a9f-8931-56da8a014a44', 'User 2'),
			('016d5895-b20f-42fd-ad6c-7f1e4c11354d', 'User 3'),
			('dc8416e1-bf15-4248-b09f-f9294adcb962', 'User 4'),
			('c1dadc8b-b0fc-41e3-a015-5a14c5c19433', 'User 5');
		
		insert into schedules (id, name, time_zone)
		values
			({{uuid "rule_sched"}}, 'Rule Test', 'UTC'),
			({{uuid "tgt_sched"}}, 'Target Test', 'UTC'),
			({{uuid "override_sched"}}, 'Override Test', 'UTC'),
			({{uuid "cal_sub_sched"}}, 'Calendar Subscriptions Test', 'UTC');

		insert into rotations (id, name, type, time_zone)
		values
			({{uuid "part_rot"}}, 'Part Rotation', 'daily', 'UTC');

		insert into user_contact_methods (id, user_id, name, type, value)
		values
			({{uuid "nr_cm"}}, {{uuid "nr_user"}}, 'Test', 'SMS', {{phone "nr"}});

		insert into escalation_policies (id, name)
		values
			({{uuid "unack_ep1"}}, 'Unack Test 1'),
			({{uuid "unack_ep2"}}, 'Unack Test 2'),
			({{uuid "int_key_ep"}}, 'Int Key Test'),
			({{uuid "hb_ep"}}, 'Heartbeat Test'),
			({{uuid "step_ep"}}, 'Step Test'),
			({{uuid "act_ep"}}, 'Action Test');

		insert into escalation_policy_steps (id, escalation_policy_id, delay)
		values
			({{uuid "act_ep_step"}}, {{uuid "act_ep"}}, 1);

		
		insert into services (id, name, escalation_policy_id)
		values
			({{uuid "int_key_svc"}}, 'Int Key Test', {{uuid "int_key_ep"}}),
			({{uuid "hb_svc"}}, 'Heartbeat Test', {{uuid "hb_ep"}}),
			({{uuid "unack_svc1"}}, 'Unack Test 1', {{uuid "unack_ep1"}}),
			({{uuid "unack_svc2"}}, 'Unack Test 2', {{uuid "unack_ep2"}});
`
	type idParser func(m map[string]interface{}) (string, bool)

	var getID idParser
	getID = func(m map[string]interface{}) (string, bool) {
		if id, ok := m["id"].(string); ok {
			return id, true
		}
		if id, ok := m["id"].(float64); ok {
			return strconv.Itoa(int(id)), true
		}
		for _, v := range m {
			if vm, ok := v.(map[string]interface{}); ok {
				id, ok := getID(vm)
				if ok {
					return id, true
				}
			}
		}
		return "", false
	}

	checkMultiInsert := func(limitID limit.ID, expErrMsg string, addQuery func(num int) string) {
		uuids := make(map[string]string)
		t.Run(string(limitID), func(t *testing.T) {
			/*
				Sequence:
				1. update to 4
				2. set limit to 2
				3. update to 5 (should fail)
				4. update to 3 (should work)
				5. update to 2 (should work)
				6. update to 3 (should fail)
				7. set limit to -1
				8. update to 4 (should work)
			*/
			t.Parallel()
			h := harness.NewHarness(t, sql, "limit-configuration")
			defer h.Close()

			doQL := func(t *testing.T, query string) (string, string) {
				t.Helper()
				g := h.GraphQLQuery2(query)
				if len(g.Errors) > 1 {
					for _, err := range g.Errors {
						t.Logf(err.Message)
					}
					t.Fatalf("got %d errors; want 0 or 1", len(g.Errors))
				}
				if len(g.Errors) == 0 {
					var m map[string]interface{}

					err := json.Unmarshal(g.Data, &m)
					if err != nil {
						t.Fatalf("got err='%s'; want nil", err.Error())
					}
					id, _ := getID(m)
					return id, ""
				}

				return "", g.Errors[0].Message
			}

			doQuery := func(t *testing.T, query string) string {
				id, errMsg := doQL(t, query)
				assert.Empty(t, errMsg, "error message")
				return id
			}
			doQueryExpectError := func(t *testing.T, query string, expErr string) {
				_, errMsg := doQL(t, query)
				t.Log(errMsg)
				t.Log(expErr)
				assert.Contains(t, errMsg, expErr, "error message")
			}

			tmplExecute := func(tmpl *template.Template, t *testing.T, qs string) string {
				tmpl, err := tmpl.Parse(qs)
				require.NoError(t, err)
				var buf bytes.Buffer
				err = tmpl.Execute(&buf, nil)
				require.NoError(t, err)
				return buf.String()
			}

			tmpl := template.New("uuids")
			tmpl.Funcs(template.FuncMap{
				"uuid": func(id string) string {
					if id == "phone" {
						p := h.Phone("")
						return p
					}
					if uuid, ok := uuids[id]; ok {
						return uuid
					}
					uuid := h.UUID(id)
					uuids[id] = uuid
					return uuid
				},
			})

			query := tmplExecute(tmpl, t, addQuery(4))
			doQuery(t, query)

			h.SetSystemLimit(limitID, 2)

			query = tmplExecute(tmpl, t, addQuery(5))
			doQueryExpectError(t, query, expErrMsg)

			query = tmplExecute(tmpl, t, addQuery(3)) // 4->3 should work
			doQuery(t, query)

			query = tmplExecute(tmpl, t, addQuery(2))
			doQuery(t, query)

			query = tmplExecute(tmpl, t, addQuery(3))
			doQueryExpectError(t, query, expErrMsg) // 2->3 should fail

			h.SetSystemLimit(limitID, -1)

			query = tmplExecute(tmpl, t, addQuery(4))
			doQuery(t, query)
		})
	}

	checkSingleInsert := func(limitID limit.ID, expErrMsg string, addQuery func(index int) string, delQuery func(ids []string) string) {
		uuids := make(map[string]string)
		t.Run(string(limitID), func(t *testing.T) {
			/*
				Sequence:
				1. create 4
				2. set limit to 2
				3. create (should fail)
				4. delete x3
				5. create (should work)
				6. create (should fail)
				7. set limit to -1
				8. create (should work)
			*/
			t.Parallel()
			h := harness.NewHarness(t, sql, "limit-configuration")
			defer h.Close()

			doQL := func(t *testing.T, query string) (string, string) {
				t.Helper()
				g := h.GraphQLQuery2(query)
				if len(g.Errors) > 1 {
					for _, err := range g.Errors {
						t.Logf(err.Message)
					}
					t.Fatalf("got %d errors; want 0 or 1", len(g.Errors))
				}
				if len(g.Errors) == 0 {
					var m map[string]interface{}

					err := json.Unmarshal(g.Data, &m)
					if err != nil {
						t.Fatalf("got err='%s'; want nil", err.Error())
					}
					id, _ := getID(m)
					return id, ""
				}

				return "", g.Errors[0].Message
			}

			doQuery := func(t *testing.T, query string) string {
				id, errMsg := doQL(t, query)
				assert.Empty(t, errMsg, "error message")
				return id
			}
			doQueryExpectError := func(t *testing.T, query string, expErr string) {
				_, errMsg := doQL(t, query)
				t.Log(errMsg)
				t.Log(expErr)
				assert.Contains(t, errMsg, expErr, "error message")
			}

			tmplExecute := func(tmpl *template.Template, t *testing.T, qs string) string {
				tmpl, err := tmpl.Parse(qs)
				require.NoError(t, err)
				var buf bytes.Buffer
				err = tmpl.Execute(&buf, nil)
				require.NoError(t, err)
				return buf.String()
			}

			tmpl := template.New("uuids")
			tmpl.Funcs(template.FuncMap{
				"uuid": func(id string) string {
					if id == "phone" {
						p := h.Phone("")
						return p
					}
					if uuid, ok := uuids[id]; ok {
						return uuid
					}
					uuid := h.UUID(id)
					uuids[id] = uuid
					return uuid
				},
			})

			// create 4
			q := make([]string, 5)
			for i := 0; i < 5; i++ {
				q[i] = tmplExecute(tmpl, t, addQuery(i))
			}

			// create 4
			ids := []string{
				doQuery(t, q[0]),
				doQuery(t, q[1]),
				doQuery(t, q[2]),
				doQuery(t, q[3]),
			}
			h.SetSystemLimit(limitID, 2)

			//create should fail
			query := tmplExecute(tmpl, t, addQuery(4))
			doQueryExpectError(t, query, expErrMsg)

			query = tmplExecute(tmpl, t, delQuery(ids))
			doQuery(t, query)

			//delQuery should always remove the first ID in the list
			ids = ids[1:]

			query = tmplExecute(tmpl, t, delQuery(ids))
			doQuery(t, query)

			ids = ids[1:]

			query = tmplExecute(tmpl, t, delQuery(ids))
			doQuery(t, query)

			//should be able to create 1 more
			query = tmplExecute(tmpl, t, addQuery(0))
			doQuery(t, query)

			//but only one
			query = tmplExecute(tmpl, t, addQuery(1))
			doQueryExpectError(t, query, expErrMsg)

			h.SetSystemLimit(limitID, -1)

			//no more limit
			query = tmplExecute(tmpl, t, addQuery(1))
			doQuery(t, query)
		})
	}

	userIDs := []string{"50322144-1e88-43dc-b638-b16a5be7bad6", "dfcc0684-f045-4a9f-8931-56da8a014a44", "016d5895-b20f-42fd-ad6c-7f1e4c11354d", "dc8416e1-bf15-4248-b09f-f9294adcb962", "c1dadc8b-b0fc-41e3-a015-5a14c5c19433"}
	var n int
	uniqName := func() string {
		n++
		return fmt.Sprintf("Thing %d", n)
	}
	s := time.Now().AddDate(1, 0, 0)
	uniqTime := func() time.Time {
		s = s.AddDate(0, 0, 1)
		return s
	}
	mapIDs := func(ids []string, fn func(string) string) string {
		if fn == nil {
			// default wrap in quotes
			fn = func(id string) string { return `"` + id + `"` }
		}
		res := make([]string, len(ids))
		for i, id := range ids {
			res[i] = fn(id)
		}
		return strings.Join(res, ",")
	}

	checkSingleInsert(
		limit.ContactMethodsPerUser,
		"contact methods",
		func(int) string {
			return fmt.Sprintf(`mutation{createUserContactMethod(input:{type: SMS, name: "%s", value: "{{uuid "phone"}}", userID: "{{uuid "cm_user"}}"}){id}}`, uniqName())
		},
		func(ids []string) string {
			return fmt.Sprintf(`mutation{deleteAll(input:[{id: "%s", type: contactMethod}])}`, ids[0])
		},
	)

	nrDelay := 0
	checkSingleInsert(
		limit.NotificationRulesPerUser,
		"notification rules",
		func(int) string {
			nrDelay++
			return fmt.Sprintf(`mutation{createUserNotificationRule(input:{contactMethodID: "{{uuid "nr_cm"}}", delayMinutes: %d, userID: "{{uuid "nr_user"}}"}){id}}`, nrDelay)
		},
		func(ids []string) string {
			return fmt.Sprintf(`mutation{deleteAll(input:[{id: "%s", type: notificationRule}])}`, ids[0])
		},
	)

	checkSingleInsert(
		limit.EPStepsPerPolicy,
		"steps",
		func(int) string {
			return fmt.Sprintf(`mutation{createEscalationPolicyStep(input:{escalationPolicyID: "{{uuid "step_ep"}}", delayMinutes: 1}){id}}`)
		},
		func(ids []string) string {
			return fmt.Sprintf(`mutation{updateEscalationPolicy(input: {id: "{{uuid "step_ep"}}", stepIDs: [%s]})}`,
				mapIDs(ids[1:], nil),
			)
		},
	)

	checkMultiInsert(
		limit.EPActionsPerStep,
		"actions",
		func(num int) string {
			return fmt.Sprintf(`mutation{updateEscalationPolicyStep(input:{id:"{{uuid "act_ep_step"}}", targets: [%s]})}`,
				mapIDs(userIDs[:num], func(id string) string { return fmt.Sprintf(`{type: user, id: "%s"}`, id) }),
			)
		},
	)

	checkMultiInsert(
		limit.ParticipantsPerRotation,
		"participants",
		func(num int) string {
			return fmt.Sprintf(`mutation{updateRotation(input:{id: "{{uuid "part_rot"}}", userIDs: [%s]})}`, mapIDs(userIDs[:num], nil))
		},
	)

	checkSingleInsert(
		limit.IntegrationKeysPerService,
		"integration keys",
		func(int) string {
			return fmt.Sprintf(`mutation{createIntegrationKey(input:{serviceID: "{{uuid "int_key_svc"}}", type: generic, name:"%s"}){id}}`, uniqName())
		},
		func(ids []string) string {
			return fmt.Sprintf(`mutation{deleteAll(input: [{id: "%s", type: integrationKey}])}`, ids[0])
		},
	)

	checkSingleInsert(
		limit.HeartbeatMonitorsPerService,
		"heartbeat monitors",
		func(int) string {
			return fmt.Sprintf(`mutation{createHeartbeatMonitor(input:{serviceID: "{{uuid "hb_svc"}}", name: "%s", timeoutMinutes: 5 }){id}}`, uniqName())
		},
		func(ids []string) string {
			return fmt.Sprintf(`mutation{deleteAll(input: [{id: "%s", type: heartbeatMonitor}])}`, ids[0])
		},
	)

	checkMultiInsert(
		limit.RulesPerSchedule,
		"rules",
		func(num int) string {
			toAdd := make([]string, num)
			return fmt.Sprintf(`mutation{updateScheduleTarget(input:{scheduleID: "{{uuid "rule_sched"}}", target: {id: "%s", type: user}, rules: [%s]})}`,
				userIDs[0],
				mapIDs(toAdd, func(string) string {
					return `{}`
				}),
			)
		},
	)

	checkSingleInsert(
		limit.TargetsPerSchedule,
		"targets",
		func(index int) string {
			return fmt.Sprintf(`mutation{updateScheduleTarget(input:{scheduleID: "{{uuid "tgt_sched"}}", target:{type:user, id: "%s"}, rules: [{}]})}`, userIDs[index])
		},
		func(ids []string) string {
			// can't use IDs since the update mutation won't return a usable ID so they will all be blank
			return fmt.Sprintf(`mutation{updateScheduleTarget(input:{scheduleID: "{{uuid "tgt_sched"}}", target:{type:user, id: "%s"}, rules: []})}`, userIDs[4-len(ids)])
		},
	)

	// ack
	checkSingleInsert(
		limit.UnackedAlertsPerService,
		"unacknowledged alerts",
		func(int) string {
			return fmt.Sprintf(`mutation{createAlert(input:{serviceID: "{{uuid "unack_svc1"}}", summary: "%s"}){id}}`, uniqName())
		},
		func(ids []string) string {
			return fmt.Sprintf(`mutation{updateAlerts(input:{alertIDs: [%s], newStatus: StatusAcknowledged}){id}}`, ids[0])
		},
	)

	// close
	checkSingleInsert(
		limit.UnackedAlertsPerService,
		"unacknowledged alerts",
		func(int) string {
			return fmt.Sprintf(`mutation{createAlert(input:{serviceID: "{{uuid "unack_svc2"}}", summary: "%s"}){id}}`, uniqName())
		},
		func(ids []string) string {
			return fmt.Sprintf(`mutation{updateAlerts(input:{alertIDs: [%s], newStatus: StatusClosed}){id}}`, ids[0])
		},
	)

	checkSingleInsert(
		limit.UserOverridesPerSchedule,
		"overrides",
		func(int) string {
			return fmt.Sprintf(`mutation{createUserOverride(input:{scheduleID: "{{uuid "override_sched"}}", addUserID: "%s", start: "%s", end: "%s"}){id}}`,
				userIDs[0],
				uniqTime().Format(time.RFC3339),
				uniqTime().Format(time.RFC3339),
			)
		},
		func(ids []string) string {
			return fmt.Sprintf(`mutation{deleteAll(input:[{type: userOverride, id: "%s"}])}`, ids[0])
		},
	)

	checkSingleInsert(
		limit.CalendarSubscriptionsPerUser,
		"subscriptions",
		func(int) string {
			return fmt.Sprintf(`
				mutation {
					createUserCalendarSubscription(
						input: {
							name: "%s"
							scheduleID: "{{uuid "cal_sub_sched"}}"
							reminderMinutes: [5, 3, 1]
							disabled: false
						}
					) { id }
				}
				`, uniqName())
		},
		func(ids []string) string {
			return fmt.Sprintf(`mutation{deleteAll(input:[{type: calendarSubscription, id: "%s"}])}`, ids[0])
		},
	)

}
