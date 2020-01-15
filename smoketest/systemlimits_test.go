package smoketest

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

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
			({{uuid "ep_act_user1"}}, 'Step 1'),
			({{uuid "ep_act_user2"}}, 'Step 2'),
			({{uuid "ep_act_user3"}}, 'Step 3'),
			({{uuid "ep_act_user4"}}, 'Step 4'),
			({{uuid "part_user"}}, 'Part User'),
			({{uuid "rule_user"}}, 'Sched Rule User'),
			({{uuid "tgt_user1"}}, 'Target 1'),
			({{uuid "tgt_user2"}}, 'Target 2'),
			({{uuid "tgt_user3"}}, 'Target 3'),
			({{uuid "tgt_user4"}}, 'Target 4');
		
		insert into schedules (id, name, time_zone)
		values
			({{uuid "rule_sched"}}, 'Rule Test', 'UTC'),
			({{uuid "tgt_sched"}}, 'Target Test', 'UTC');
		
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
			({{uuid "act_ep"}}, 'Action Test'),
			({{uuid "act_ep2"}}, 'Action Test 2');

		insert into escalation_policy_steps (id, escalation_policy_id, delay)
		values
			({{uuid "act_ep_step"}}, {{uuid "act_ep"}}, 15),
			({{uuid "act_ep_step2"}}, {{uuid "act_ep2"}}, 15);
		
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

	doQLErr := func(t *testing.T, query string, getID idParser) (string, string) {
		g := h.GraphQLQuery2(query)
		errs := len(g.Errors)
		if errs > 1 {
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

	doTest := func(limitID limit.ID, expErrMsg string, addQuery func(int) string, delQuery func(int, string) string, parseID idParser) {
		if parseID == nil {
			parseID = getID
		}
		t.Run(string(limitID), func(t *testing.T) {
			/*
				Sequence:
				1. create 3
				2. set limit to 2
				3. create (should fail)
				4. delete x2
				5. create (should work)
				6. create (should fail)
				7. set limit to -1
				8. create (should work)
			*/
			noErr := func(id, res string) string {
				t.Helper()
				if res == "" {
					return id
				}
				t.Fatalf("got err='%s'; want nil", res)
				panic("test did not abort")
			}
			mustErr := func(id, res string) {
				t.Helper()
				if !strings.Contains(res, expErrMsg) {
					t.Fatalf("err='%s'; want substring '%s'", res, expErrMsg)
				}
			}
			setLimit := func(max int) {
				t.Helper()
				h.SetSystemLimit(limitID, max)
			}
			ids := []string{ // create 3
				noErr(doQLErr(t, addQuery(1), parseID)),
				noErr(doQLErr(t, addQuery(2), parseID)),
				noErr(doQLErr(t, addQuery(3), parseID)),
			}
			setLimit(2)                                     // set limit to 2
			mustErr(doQLErr(t, addQuery(4), parseID))       // create should fail
			noErr(doQLErr(t, delQuery(2, ids[2]), parseID)) // delete 2
			noErr(doQLErr(t, delQuery(1, ids[1]), parseID))

			noErr(doQLErr(t, addQuery(2), parseID))   // should be able to create 1 more
			mustErr(doQLErr(t, addQuery(3), parseID)) // but only one

			setLimit(-1)

			noErr(doQLErr(t, addQuery(3), parseID)) // no more limit
		})
	}

	var n int
	name := func() string {
		n++
		return fmt.Sprintf("Thing %d", n)
	}

	doTest(
		limit.ContactMethodsPerUser,
		"contact methods",
		func(int) string {
			return fmt.Sprintf(`mutation{createUserContactMethod(input:{type: SMS, name: "%s", value: "%s", userID: "%s"}){id}}`, name(), h.Phone(""), h.UUID("cm_user"))
		},
		func(_ int, id string) string {
			return fmt.Sprintf(`mutation{deleteAll(input:[{id: "%s", type: contactMethod}])}`, id)
		},
		nil,
	)

	nrDelay := 0
	doTest(
		limit.NotificationRulesPerUser,
		"notification rules",
		func(int) string {
			nrDelay++
			return fmt.Sprintf(`mutation{createUserNotificationRule(input:{contactMethodID: "%s", delayMinutes: %d, userID: "%s"}){id}}`, h.UUID("nr_cm"), nrDelay, h.UUID("nr_user"))
		},
		func(_ int, id string) string {
			return fmt.Sprintf(`mutation{deleteAll(input:[{id: "%s", type: notificationRule}])}`, id)
		},
		nil,
	)

	// doTest(
	// 	limit.EPStepsPerPolicy,
	// 	"steps",
	// 	func(int) string {
	// 		return fmt.Sprintf(`mutation{createOrUpdateEscalationPolicyStep(input:{escalation_policy_id: "%s", delay_minutes: 10, user_ids: [], schedule_ids: []}){escalation_policy_step{id}}}`, h.UUID("step_ep"))
	// 	},
	// 	func(_ int, id string) string {
	// 		return fmt.Sprintf(`mutation{deleteEscalationPolicyStep(input:{id:"%s"}){id: deleted_id}}`, id)
	// 	},
	// 	nil,
	// )

	// actUsersList := []string{h.UUID("ep_act_user1"), h.UUID("ep_act_user2"), h.UUID("ep_act_user3"), h.UUID("ep_act_user4")}
	// doTest(
	// 	limit.EPActionsPerStep,
	// 	"actions",
	// 	func(n int) string {
	// 		return fmt.Sprintf(`mutation{addEscalationPolicyStepTarget(input:{step_id:"%s", target_id: "%s", target_type: user}){id: target_id}}`,
	// 			h.UUID("act_ep_step2"),
	// 			actUsersList[n-1],
	// 		)
	// 	},
	// 	func(_ int, id string) string {
	// 		return fmt.Sprintf(`mutation{deleteEscalationPolicyStepTarget(input:{step_id:"%s", target_id: "%s", target_type: user}){id: target_id}}`,
	// 			h.UUID("act_ep_step2"),
	// 			id,
	// 		)
	// 	},
	// 	nil,
	// )

	// doTest(
	// 	limit.ParticipantsPerRotation,
	// 	"participants",
	// 	func(int) string {
	// 		return fmt.Sprintf(`mutation{addRotationParticipant(input:{user_id: "%s", rotation_id: "%s"}){id}}`, h.UUID("part_user"), h.UUID("part_rot"))
	// 	},
	// 	func(_ int, id string) string {
	// 		return fmt.Sprintf(`mutation{deleteRotationParticipant(input:{id: "%s"}){id: deleted_id}}`, id)
	// 	},
	// 	nil,
	// )

	// doTest(
	// 	limit.IntegrationKeysPerService,
	// 	"integration keys",
	// 	func(int) string {
	// 		return fmt.Sprintf(`mutation{createIntegrationKey(input:{service_id: "%s", type: generic, name:"%s"}){id}}`, h.UUID("int_key_svc"), name())
	// 	},
	// 	func(_ int, id string) string {
	// 		return fmt.Sprintf(`mutation{deleteIntegrationKey(input:{id: "%s"}){id: deleted_id}}`, id)
	// 	},
	// 	nil,
	// )

	// doTest(
	// 	limit.HeartbeatMonitorsPerService,
	// 	"heartbeat monitors",
	// 	func(int) string {
	// 		return fmt.Sprintf(`mutation{createAll(input:{heartbeat_monitors: [{interval_minutes:5,service_id: "%s", name: "%s"}]}){heartbeat_monitors {id}}}`, h.UUID("hb_svc"), name())
	// 	},
	// 	func(_ int, id string) string {
	// 		return fmt.Sprintf(`mutation{deleteHeartbeatMonitor(input:{id: "%s"}){id: deleted_id}}`, id)
	// 	},
	// 	func(m map[string]interface{}) (string, bool) {
	// 		c, ok := m["createAll"].(map[string]interface{})
	// 		if !ok {
	// 			return "", false
	// 		}
	// 		asn := c["heartbeat_monitors"].([]interface{})
	// 		return asn[0].(map[string]interface{})["id"].(string), true
	// 	},
	// )

	// // schedule tests (need custom parser)
	// s := time.Date(2005, 0, 0, 0, 0, 0, 0, time.UTC)
	// startTime := func() string {
	// 	s = s.Add(time.Minute)
	// 	return s.Format("15:04")
	// }
	// doTest(
	// 	limit.RulesPerSchedule,
	// 	"rules",
	// 	func(int) string {
	// 		return fmt.Sprintf(`mutation{createScheduleRule(input:{
	// 			schedule_id: "%s",
	// 			target_id: "%s",
	// 			target_type: user,
	// 			sunday: false,monday: false,tuesday: false,wednesday: false,thursday: false,friday: false,saturday: false,
	// 			start: "%s",
	// 			end: "%s"
	// 		}){assignments(start_time: "%s", end_time: "%s"){rules{id, start}}}}`,
	// 			h.UUID("rule_sched"),
	// 			h.UUID("rule_user"),
	// 			startTime(),
	// 			startTime(),
	// 			s.Format(time.RFC3339),
	// 			s.Format(time.RFC3339),
	// 		)
	// 	},
	// 	func(_ int, id string) string {
	// 		return fmt.Sprintf(`mutation{deleteScheduleRule(input:{id: "%s"}){id}}`, id)
	// 	},
	// 	func(m map[string]interface{}) (string, bool) {
	// 		sched, ok := m["createScheduleRule"].(map[string]interface{})
	// 		if !ok {
	// 			return "", false
	// 		}
	// 		asn := sched["assignments"].([]interface{})
	// 		rules := asn[0].(map[string]interface{})["rules"].([]interface{})
	// 		sort.Slice(rules, func(i, j int) bool {
	// 			return rules[i].(map[string]interface{})["start"].(string) < rules[j].(map[string]interface{})["start"].(string)
	// 		})
	// 		return rules[len(rules)-1].(map[string]interface{})["id"].(string), true
	// 	},
	// )

	// tgtUsersList := []string{h.UUID("tgt_user1"), h.UUID("tgt_user2"), h.UUID("tgt_user3"), h.UUID("tgt_user4")}
	// doTest(
	// 	limit.TargetsPerSchedule,
	// 	"targets",
	// 	func(n int) string {
	// 		return fmt.Sprintf(`mutation{createScheduleRule(input:{
	// 			schedule_id: "%s",
	// 			target_id: "%s",
	// 			target_type: user,
	// 			sunday: false,monday: false,tuesday: false,wednesday: false,thursday: false,friday: false,saturday: false,
	// 			start: "%s",
	// 			end: "%s"
	// 		}){assignments(start_time: "%s", end_time: "%s"){rules{id, start}}}}`,
	// 			h.UUID("tgt_sched"),
	// 			tgtUsersList[n-1],
	// 			startTime(),
	// 			startTime(),
	// 			s.Format(time.RFC3339),
	// 			s.Format(time.RFC3339),
	// 		)
	// 	},
	// 	func(_ int, id string) string {
	// 		return fmt.Sprintf(`mutation{deleteScheduleRule(input:{id: "%s"}){id}}`, id)
	// 	},
	// 	func(m map[string]interface{}) (string, bool) {
	// 		sched, ok := m["createScheduleRule"].(map[string]interface{})
	// 		if !ok {
	// 			return "", false
	// 		}
	// 		asn := sched["assignments"].([]interface{})
	// 		var rules []interface{}
	// 		for _, a := range asn {
	// 			rules = append(rules, a.(map[string]interface{})["rules"].([]interface{})...)
	// 		}
	// 		sort.Slice(rules, func(i, j int) bool {
	// 			return rules[i].(map[string]interface{})["start"].(string) < rules[j].(map[string]interface{})["start"].(string)
	// 		})
	// 		return rules[len(rules)-1].(map[string]interface{})["id"].(string), true
	// 	},
	// )

	// doTest(
	// 	limit.UnackedAlertsPerService,
	// 	"unacknowledged alerts",
	// 	func(int) string {
	// 		return fmt.Sprintf(`mutation{createAlert(input:{service_id: "%s", description: "%s"}){id: _id}}`, h.UUID("unack_svc1"), name())
	// 	},
	// 	func(_ int, id string) string {
	// 		return fmt.Sprintf(`mutation{updateAlertStatus(input:{id:%s, status: acknowledged}){id}}`, id)
	// 	},
	// 	nil,
	// )
	// doTest(
	// 	limit.UnackedAlertsPerService,
	// 	"unacknowledged alerts",
	// 	func(int) string {
	// 		return fmt.Sprintf(`mutation{createAlert(input:{service_id: "%s", description: "%s"}){id: _id}}`, h.UUID("unack_svc2"), name())
	// 	},
	// 	func(_ int, id string) string {
	// 		return fmt.Sprintf(`mutation{updateAlertStatus(input:{id:%s, status: closed}){id}}`, id)
	// 	},
	// 	nil,
	// )

}
