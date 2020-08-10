package smoketest

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

/*
 * This file tests that the alerts GraphQL query shows the proper amount
 * of results when flipping between "includeNotified" and "favoritesOnly"
 * query options.
 *
 * Service 1: Notified alert
 * Service 2: Favorited alert
 */
func TestNotifiedAlerts(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email, role) 
	values ({{uuid "user"}}, 'bob', 'joe', 'user');

	insert into escalation_policies (id, name) 
	values ({{uuid "eid2"}}, 'esc policy 2');

	insert into services (id, escalation_policy_id, name) 
	values ({{uuid "sid2"}}, {{uuid "eid2"}}, 'service 2');
	`

	h := harness.NewHarness(t, sql, "prometheus-alertmanager-integration")
	defer h.Close()

	doQL := func(t *testing.T, h *harness.Harness, query string, res interface{}) {
		t.Helper()
		g := h.GraphQLQueryUserT(t, h.UUID("user"), query)
		// g := h.GraphQLQuery2(query)
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

	userID := h.UUID("user")
	phone := h.Phone("1")

	// create bob's contact method
	var cm struct {
		CreateUserContactMethod struct {
			ID string
		}
	}

	doQL(t, h, fmt.Sprintf(`
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
	`, userID, phone), &cm)

	// verify bob's contact method
	doQL(t, h, fmt.Sprintf(`
		mutation {
			sendContactMethodVerification(input:{
				contactMethodID: "%s"
			})
		}
	`, cm.CreateUserContactMethod.ID), nil)
	msg := h.Twilio(t).Device(phone).ExpectSMS("verification")
	digits := func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}
	codeStr := strings.Map(digits, msg.Body())
	code, _ := strconv.Atoi(codeStr)
	doQL(t, h, fmt.Sprintf(`
		mutation {
			verifyContactMethod(input:{
				contactMethodID:  "%s",
				code: %d
			})
		}
	`, cm.CreateUserContactMethod.ID, code), nil)

	// create bob's notification rule
	doQL(t, h, fmt.Sprintf(`
		mutation {
			createUserNotificationRule(input:{
				userID: "%s",
				contactMethodID: "%s",
				delayMinutes: 0
			}){
				id
			}
		}

	`, userID, cm.CreateUserContactMethod.ID), nil)

	// create escalation policy with bob on a step
	var esc struct{ CreateEscalationPolicy struct{ ID string } }
	doQL(t, h, `
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
	doQL(t, h, fmt.Sprintf(`
		mutation {
			createEscalationPolicyStep(input:{
				delayMinutes: 60,
				escalationPolicyID: "%s",
				targets: [{id: "%s", type: user}]
			}){
				id
			}
		}
	`, esc.CreateEscalationPolicy.ID, userID), &step)

	// create service for notified alert
	var svc struct{ CreateService struct{ ID string } }
	doQL(t, h, fmt.Sprintf(`
		mutation {
			createService(input:{
				name: "service 1",
				escalationPolicyID: "%s"
			}){id}
		}
	`, esc.CreateEscalationPolicy.ID), &svc)

	// favorite the second service from initSQL
	doQL(t, h, fmt.Sprintf(`
		mutation{
			setFavorite(input:{
				target:{
					id: "%s"
					type: service
				}
				favorite: true
			})
		}
	`, h.UUID("sid2")), nil)

	var s struct {
		Service struct {
			IsFavorite bool
		}
	}

	// assert the second service was favorited
	doQL(t, h, fmt.Sprintf(`
		query {
			service(id: "%s") { 
				isFavorite 
			}	
		}
	`, h.UUID("sid2")), &s)
	if s.Service.IsFavorite != true {
		t.Fatalf("ERROR: ServiceID %s IsUserFavorite=%t; want true", h.UUID("sid2"), s.Service.IsFavorite)
	}

	// create alerts against both services (notifed version & favorited version)
	h.CreateAlert(svc.CreateService.ID, "notified alert")
	h.CreateAlert(h.UUID("sid2"), "favorited alert")

	type Alerts struct {
		Alerts struct {
			Nodes []struct {
				ID string
			}
		}
	}

	var alerts1, alerts2, alerts3 Alerts

	// query for only favorites: 1 alert (the favorited one)
	doQL(t, h, `query {
		alerts(input: {
			includeNotified: false
			favoritesOnly: true
		}) {
			nodes {
				id
				summary
			}
		}
	}`, &alerts1)

	if len(alerts1.Alerts.Nodes) != 1 {
		t.Errorf("got %d alerts; want 1", len(alerts1.Alerts.Nodes))
	}

	// query for favorites & notified: 2 alerts
	doQL(t, h, `query {
			alerts(input: {
				includeNotified: true
				favoritesOnly: true
			}) {
				nodes {
					id
					summary
				}
			}
		}`, &alerts2)

	if len(alerts2.Alerts.Nodes) != 2 {
		t.Errorf("got %d alerts; want 2", len(alerts2.Alerts.Nodes))
	}

	// All Services test (favoritesOnly: false)
	// output: 2 alerts
	doQL(t, h, `query {
		alerts(input: {
			includeNotified: true
			favoritesOnly: false
		}) {
			nodes {
				id
				summary
			}
		}
	}`, &alerts3)

	if len(alerts3.Alerts.Nodes) != 2 {
		t.Errorf("got %d alerts; want 2", len(alerts3.Alerts.Nodes))
	}

	// Expect 1 SMS for the created alert against bob's CM
	h.Twilio(t).Device(phone).ExpectSMS("alert1")
}
