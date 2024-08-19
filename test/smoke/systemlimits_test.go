package smoke

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/limit"
	"github.com/target/goalert/test/smoke/harness"
)

// TestSystemLimits tests that limits are enforced if configured.
func TestSystemLimits(t *testing.T) {
	t.Parallel()

	const sql = `
		insert into users (id, name)
		values
			({{uuid "cm_user"}}, 'CM User'),
			({{uuid "nr_user"}}, 'NR User'),
			({{uuid "generic_user1"}}, 'User 1'),
			({{uuid "generic_user2"}}, 'User 2'),
			({{uuid "generic_user3"}}, 'User 3'),
			({{uuid "generic_user4"}}, 'User 4'),
			({{uuid "generic_user5"}}, 'User 5');
		
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

	h := harness.NewHarness(t, sql, "limit-configuration")
	defer h.Close()

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

	doQL := func(t *testing.T, query string) (string, string) {
		t.Helper()
		g := h.GraphQLQuery2(query)
		if len(g.Errors) > 1 {
			for _, err := range g.Errors {
				t.Log(err.Message)
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
	doQueryExpectError := func(t *testing.T, query, expErr string) {
		_, errMsg := doQL(t, query)
		assert.Contains(t, errMsg, expErr, "error message")
	}

	checkMultiInsert := func(limitID limit.ID, expErrMsg string, addQuery func(num int) string) {
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
			doQuery(t, addQuery(4))
			h.SetSystemLimit(limitID, 2)
			doQueryExpectError(t, addQuery(5), expErrMsg)
			doQuery(t, addQuery(3)) // 4->3 should work
			doQuery(t, addQuery(2))
			doQueryExpectError(t, addQuery(3), expErrMsg) // 2->3 should fail
			h.SetSystemLimit(limitID, -1)
			doQuery(t, addQuery(4))
		})
	}

	checkSingleInsert := func(limitID limit.ID, expErrMsg string, addQuery func(index int) string, delQuery func(ids []string) string) {
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

			ids := []string{ // create 4
				doQuery(t, addQuery(0)),
				doQuery(t, addQuery(1)),
				doQuery(t, addQuery(2)),
				doQuery(t, addQuery(3)),
			}
			h.SetSystemLimit(limitID, 2)
			doQueryExpectError(t, addQuery(4), expErrMsg) // create should fail
			doQuery(t, delQuery(ids))
			ids = ids[1:] // delQuery should always remove the first ID in the list
			doQuery(t, delQuery(ids))
			ids = ids[1:]
			doQuery(t, delQuery(ids))

			doQuery(t, addQuery(0))                       // should be able to create 1 more
			doQueryExpectError(t, addQuery(1), expErrMsg) // but only one

			h.SetSystemLimit(limitID, -1)

			doQuery(t, addQuery(1)) // no more limit
		})
	}

	userIDs := []string{h.UUID("generic_user1"), h.UUID("generic_user2"), h.UUID("generic_user3"), h.UUID("generic_user4"), h.UUID("generic_user5")}
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
			return fmt.Sprintf(`mutation{createUserContactMethod(input:{type: SMS, name: "%s", value: "%s", userID: "%s"}){id}}`, uniqName(), h.Phone(""), h.UUID("cm_user"))
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
			return fmt.Sprintf(`mutation{createUserNotificationRule(input:{contactMethodID: "%s", delayMinutes: %d, userID: "%s"}){id}}`, h.UUID("nr_cm"), nrDelay, h.UUID("nr_user"))
		},
		func(ids []string) string {
			return fmt.Sprintf(`mutation{deleteAll(input:[{id: "%s", type: notificationRule}])}`, ids[0])
		},
	)

	checkSingleInsert(
		limit.EPStepsPerPolicy,
		"steps",
		func(int) string {
			return fmt.Sprintf(`mutation{createEscalationPolicyStep(input:{escalationPolicyID: "%s", delayMinutes: 1}){id}}`, h.UUID("step_ep"))
		},
		func(ids []string) string {
			return fmt.Sprintf(`mutation{updateEscalationPolicy(input: {id: "%s", stepIDs: [%s]})}`,
				h.UUID("step_ep"),
				mapIDs(ids[1:], nil),
			)
		},
	)

	checkMultiInsert(
		limit.EPActionsPerStep,
		"actions",
		func(num int) string {
			return fmt.Sprintf(`mutation{updateEscalationPolicyStep(input:{id:"%s", targets: [%s]})}`,
				h.UUID("act_ep_step"),
				mapIDs(userIDs[:num], func(id string) string { return fmt.Sprintf(`{type: user, id: "%s"}`, id) }),
			)
		},
	)

	checkMultiInsert(
		limit.ParticipantsPerRotation,
		"participants",
		func(num int) string {
			return fmt.Sprintf(`mutation{updateRotation(input:{id: "%s", userIDs: [%s]})}`, h.UUID("part_rot"), mapIDs(userIDs[:num], nil))
		},
	)

	checkSingleInsert(
		limit.IntegrationKeysPerService,
		"integration keys",
		func(int) string {
			return fmt.Sprintf(`mutation{createIntegrationKey(input:{serviceID: "%s", type: generic, name:"%s"}){id}}`, h.UUID("int_key_svc"), uniqName())
		},
		func(ids []string) string {
			return fmt.Sprintf(`mutation{deleteAll(input: [{id: "%s", type: integrationKey}])}`, ids[0])
		},
	)

	checkSingleInsert(
		limit.HeartbeatMonitorsPerService,
		"heartbeat monitors",
		func(int) string {
			return fmt.Sprintf(`mutation{createHeartbeatMonitor(input:{serviceID: "%s", name: "%s", timeoutMinutes: 5 }){id}}`, h.UUID("hb_svc"), uniqName())
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
			return fmt.Sprintf(`mutation{updateScheduleTarget(input:{scheduleID: "%s", target: {id: "%s", type: user}, rules: [%s]})}`,
				h.UUID("rule_sched"),
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
			return fmt.Sprintf(`mutation{updateScheduleTarget(input:{scheduleID: "%s", target:{type:user, id: "%s"}, rules: [{}]})}`, h.UUID("tgt_sched"), userIDs[index])
		},
		func(ids []string) string {
			// can't use IDs since the update mutation won't return a usable ID so they will all be blank
			return fmt.Sprintf(`mutation{updateScheduleTarget(input:{scheduleID: "%s", target:{type:user, id: "%s"}, rules: []})}`, h.UUID("tgt_sched"), userIDs[4-len(ids)])
		},
	)

	// ack
	checkSingleInsert(
		limit.UnackedAlertsPerService,
		"unacknowledged alerts",
		func(int) string {
			return fmt.Sprintf(`mutation{createAlert(input:{serviceID: "%s", summary: "%s"}){id}}`, h.UUID("unack_svc1"), uniqName())
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
			return fmt.Sprintf(`mutation{createAlert(input:{serviceID: "%s", summary: "%s"}){id}}`, h.UUID("unack_svc2"), uniqName())
		},
		func(ids []string) string {
			return fmt.Sprintf(`mutation{updateAlerts(input:{alertIDs: [%s], newStatus: StatusClosed}){id}}`, ids[0])
		},
	)

	checkSingleInsert(
		limit.UserOverridesPerSchedule,
		"overrides",
		func(int) string {
			return fmt.Sprintf(`mutation{createUserOverride(input:{scheduleID: "%s", addUserID: "%s", start: "%s", end: "%s"}){id}}`,
				h.UUID("override_sched"),
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
						scheduleID: "%s"
						reminderMinutes: [5, 3, 1]
						disabled: false
					}
				) { id }
			}
			`, uniqName(), h.UUID("cal_sub_sched"))
		},
		func(ids []string) string {
			return fmt.Sprintf(`mutation{deleteAll(input:[{type: calendarSubscription, id: "%s"}])}`, ids[0])
		},
	)
}
