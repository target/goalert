package smoke

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/test/smoke/harness"
)

// TestUserVerificationCleanup tests that verification codes are cleaned up
// when they are expired. Also tests that contact methods that are not verified
// are cleaned up with the verification code.
func TestUserVerificationCleanup(t *testing.T) {
	t.Parallel()

	sql := `
		insert into users (id, name, email) 
		values 
			({{uuid "user"}}, 'bob', 'joe');
		insert into user_contact_methods (id, user_id, name, type, value, disabled) 
		values
				({{uuid "cm1"}}, {{uuid "user"}}, 'personal1', 'SMS', {{phone "1"}}, false),
				({{uuid "cm2"}}, {{uuid "user"}}, 'personal2', 'SMS', {{phone "2"}}, true);
	`
	h := harness.NewHarness(t, sql, "")
	defer h.Close()

	type cm struct {
		ID       string `json:"id"`
		Disabled bool   `json:"disabled"`
		Pending  bool   `json:"pending"`
	}

	type cmCreate struct {
		CreateUserContactMethod cm
	}

	type data struct {
		User struct {
			ID             string `json:"id"`
			ContactMethods []cm   `json:"contactMethods"`
		} `json:"user"`
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

	createCM := func(t *testing.T, user, name, phone string) (cm *cmCreate) {
		t.Helper()
		doQL(t, h, fmt.Sprintf(`
			mutation {
				createUserContactMethod(input: {
					userID: "%s",
					type: SMS,
					name: "%s",
					value: "%s"
				}) {
					id
				}
			}
		`, user, name, phone), &cm)
		return
	}
	cm3 := createCM(t, h.UUID("user"), "personal3", h.Phone("3"))
	cm4 := createCM(t, h.UUID("user"), "personal4", h.Phone("4"))

	verifyCM := func(t *testing.T, cmID string) {
		t.Helper()
		doQL(t, h, fmt.Sprintf(`
			mutation {
				sendContactMethodVerification(input: {
					contactMethodID: "%s"
				})
			}
		`, cmID), nil)
	}
	verifyCM(t, cm3.CreateUserContactMethod.ID)
	verifyCM(t, cm4.CreateUserContactMethod.ID)

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("3"))
	d2 := tw.Device(h.Phone("4"))

	d1Msg := d1.ExpectSMS("verification")
	d2.ExpectSMS("verification")

	code := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, d1Msg.Body())

	doQL(t, h, fmt.Sprintf(`
		mutation {
			verifyContactMethod(input: {
				contactMethodID: "%s",
				code: %s
			})
		}
	`, cm3.CreateUserContactMethod.ID, code), nil)

	h.FastForward(20 * time.Minute)
	h.Trigger()

	var d data
	doQL(t, h, fmt.Sprintf(`
			query {
				user(id: "%s") {
					id
					contactMethods {
						id
						disabled
						pending
					}
				}
			}
		`, h.UUID("user")), &d)
	assert.Len(t, d.User.ContactMethods, 3)

	expectedCMs := []string{h.UUID("cm1"), h.UUID("cm2"), cm3.CreateUserContactMethod.ID}
	sort.Strings(expectedCMs)
	var actualCMs []string
	for _, cm := range d.User.ContactMethods {
		actualCMs = append(actualCMs, cm.ID)
	}
	sort.Strings(actualCMs)
	assert.Equal(t, expectedCMs, actualCMs)
}
