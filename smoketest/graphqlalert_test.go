package smoketest

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/target/goalert/smoketest/harness"
)

// TestGraphQLAlert tests that all steps up to, and including, generating
// an alert via GraphQL result in notifications going out.
//
// Specifically, mutations tested include:
// - createContactMethod
// - createNotificationRule
// - createSchedule
// - updateSchedule
// - addRotationParticipant
// - createEscalationPolicy
// - createEscalationPolicyStep
// - createService
// - createAlert
func TestGraphQLAlert(t *testing.T) {
	t.Parallel()

	const sql = `
	insert into users (id, name, email)
	values
		({{uuid "u1"}}, 'bob', 'joe'),
		({{uuid "u2"}}, 'ben', 'josh');
`
	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		t.Fatal("failed to load America/Chicago tzdata:", err)
	}

	h := harness.NewHarness(t, sql, "ids-to-uuids")
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

	uid1, uid2 := h.UUID("u1"), h.UUID("u2")
	phone1, phone2 := h.Phone("u1"), h.Phone("u2")

	var cm1, cm2 struct {
		CreateUserContactMethod struct {
			ID string `json:"id"`
		} `json:"createUserContactMethod"`
	}
	doQL(fmt.Sprintf(`
		mutation {
			createUserContactMethod(input:{
				userID: "%s",
				name: "default",
				type: SMS,
				value: "%s"
			}) {
				id
			}
		}
	`, uid1, phone1), &cm1)
	doQL(fmt.Sprintf(`
		mutation {
			createUserContactMethod(input:{
				userID: "%s",
				name: "default",
				type: SMS,
				value: "%s"
			}) {
				id
			}
		}
	`, uid2, phone2), &cm2)

	doQL(fmt.Sprintf(`
		mutation {
			sendContactMethodVerification(input:{
				contactMethodID: "%s"
			})
		}
	`, cm1.CreateUserContactMethod.ID), nil)

	doQL(fmt.Sprintf(`
		mutation {
			sendContactMethodVerification(input:{
				contactMethodID: "%s"
			})
		}
	`, cm2.CreateUserContactMethod.ID), nil)

	tw := h.Twilio()
	d1 := tw.Device(phone1)
	d2 := tw.Device(phone2)

	msg1 := d1.ExpectSMS("verification")
	msg2 := d2.ExpectSMS("verification")
	tw.WaitAndAssert() // wait for code, and ensure no notifications went out

	codeStr1 := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, msg1.Body())

	codeStr2 := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, msg2.Body())

	code1, _ := strconv.Atoi(codeStr1)
	code2, _ := strconv.Atoi(codeStr2)

	doQL(fmt.Sprintf(`
		mutation {
			verifyContactMethod(input:{
				contactMethodID:  "%s",
				code: %d
			})
		}
	`, cm1.CreateUserContactMethod.ID, code1), nil)

	doQL(fmt.Sprintf(`
		mutation {
			verifyContactMethod(input:{
				contactMethodID:  "%s",
				code: %d
			})
		}
	`, cm2.CreateUserContactMethod.ID, code2), nil)

	doQL(fmt.Sprintf(`
		mutation {
			createUserNotificationRule(input:{
				userID: "%s",
				contactMethodID: "%s",
				delayMinutes: 0
			}){
				id
			}
		}
	
	`, uid1, cm1.CreateUserContactMethod.ID), nil)

	doQL(fmt.Sprintf(`
		mutation {
			createUserNotificationRule(input:{
				userID: "%s",
				contactMethodID: "%s",
				delayMinutes: 0
			}){
				id
			}
		}
	
	`, uid2, cm2.CreateUserContactMethod.ID), nil)

	var sched struct {
		CreateSchedule struct {
			ID      string
			Name    string
			Targets []struct {
				ScheduleID string
				Target     struct{ ID string }
			}
		}
	}

	doQL(fmt.Sprintf(`
		mutation {
			createSchedule(
				input: {
					name: "default"
					description: "default testing"
					timeZone: "America/Chicago"
					targets: {
						newRotation: {
							name: "foobar"
							timeZone: "America/Chicago"
							start: "%s"
							type: daily
						}
						rules: {
							start: "00:00"
							end: "23:00"
							weekdayFilter: [true, true, true, true, true]
						}
					}
				}
			) {
				id
				name
				targets {
					scheduleID
					target {
						id
					}
				}
			}
		}
	`, time.Now().Add(-time.Hour).In(loc).Format(time.RFC3339)), &sched)

	if len(sched.CreateSchedule.Targets) != 1 {
		t.Errorf("got %d schedule targets; want 1", len(sched.CreateSchedule.Targets))
	}

	rotID := sched.CreateSchedule.Targets[0].Target.ID

	doQL(fmt.Sprintf(`
		mutation {
			updateRotation(input:{
				id: "%s",
				userIDs: ["%s"]
			})
		}
	
	`, rotID, uid1), nil)

	var esc struct{ CreateEscalationPolicy struct{ ID string } }
	doQL(`
		mutation {
			createEscalationPolicy(input:{
				repeat: 0,
				name: "default"
			}){id}
		}
	`, &esc)

	var step struct {
		CreateEscalationPolicyStep struct{ Step struct{ ID string } }
	}
	doQL(fmt.Sprintf(`
		mutation {
			createEscalationPolicyStep(input:{
				delayMinutes: 60,
				escalationPolicyID: "%s",
				targets: [{id: "%s", type: user}, {id: "%s", type: schedule}]
			}){
				id
			}
		}
	`, esc.CreateEscalationPolicy.ID, uid2, sched.CreateSchedule.ID), &step)
	var svc struct{ CreateService struct{ ID string } }
	doQL(fmt.Sprintf(`
		mutation {
			createService(input:{
				name: "default",
				escalationPolicyID: "%s"
			}){id}
		}
	`, esc.CreateEscalationPolicy.ID), &svc)

	// finally.. we can create the alert
	doQL(fmt.Sprintf(`
		mutation {
			createAlert(input:{
				summary: "brok",
				serviceID: "%s"
			}){id}
		}
	`, svc.CreateService.ID), nil)

	h.Twilio().Device(phone1).ExpectSMS()
	h.Twilio().Device(phone2).ExpectSMS()
}
