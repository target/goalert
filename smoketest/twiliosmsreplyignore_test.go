package smoketest

import (
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

// TestTwilioSMSReplyIgnore checks that replies are dropped/ignored for unknown/disabled contact methods.
func TestTwilioSMSReplyIgnore(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email, role) 
	values 
		({{uuid "user"}}, 'bob', 'joe', 'user');
	insert into user_contact_methods (id, disabled, user_id, name, type, value) 
	values
		({{uuid "cm1"}}, false, {{uuid "user"}}, 'personal1', 'SMS', {{phone "1"}}),
		({{uuid "cm2"}}, true, {{uuid "user"}}, 'personal2', 'SMS', {{phone "2"}});
`

	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))
	d2 := tw.Device(h.Phone("2"))
	d3 := tw.Device(h.Phone("3"))

	d1.SendSMS("nonsense")
	d2.SendSMS("nonsense")
	d3.SendSMS("nonsense")

	d1.ExpectSMS("sorry")

	// d2 should not get a message, as it's disabled
	// d3 should not get a message, as it's unknown

}
