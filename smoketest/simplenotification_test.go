package smoketest

import (
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

// TestSimpleNotifications tests that SMS and Voice goes out for
// 1 alert -> service -> esc -> step -> user. 2 rules (1 of each) immediately.
func TestSimpleNotifications(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "user"}}, 'bob', 'joe');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}}),
		({{uuid "cm2"}}, {{uuid "user"}}, 'personal', 'VOICE', {{phone "1"}}),
		({{uuid "cm3"}}, {{uuid "user"}}, 'personal1', 'SMS', {{phoneCC "+91" "1"}}),
		({{uuid "cm4"}}, {{uuid "user"}}, 'personal1', 'VOICE', {{phoneCC "+91" "1"}}),
		({{uuid "cm5"}}, {{uuid "user"}}, 'personal2', 'SMS', {{phoneCC "+49" "1"}}),
		({{uuid "cm6"}}, {{uuid "user"}}, 'personal2', 'VOICE', {{phoneCC "+49" "1"}}),
		({{uuid "cm7"}}, {{uuid "user"}}, 'personal3', 'SMS', {{phoneCC "+44" "1"}}),
		({{uuid "cm8"}}, {{uuid "user"}}, 'personal3', 'VOICE', {{phoneCC "+44" "1"}}),
		({{uuid "cm9"}}, {{uuid "user"}}, 'personal4', 'SMS', {{phoneCC "+852" "1"}}),
		({{uuid "cm10"}}, {{uuid "user"}}, 'personal4', 'VOICE', {{phoneCC "+852" "1"}}),
		({{uuid "cm11"}}, {{uuid "user"}}, 'personal5', 'SMS', {{phoneCC "+86" "1"}}),
		({{uuid "cm12"}}, {{uuid "user"}}, 'personal5', 'VOICE', {{phoneCC "+86" "1"}}),
		({{uuid "cm13"}}, {{uuid "user"}}, 'personal6', 'SMS', {{phoneCC "+502" "1"}}),
		({{uuid "cm14"}}, {{uuid "user"}}, 'personal6', 'VOICE', {{phoneCC "+502" "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "user"}}, {{uuid "cm1"}}, 0),
		({{uuid "user"}}, {{uuid "cm2"}}, 0),
		({{uuid "user"}}, {{uuid "cm3"}}, 0),
		({{uuid "user"}}, {{uuid "cm4"}}, 0),
		({{uuid "user"}}, {{uuid "cm5"}}, 0),
		({{uuid "user"}}, {{uuid "cm6"}}, 0),
		({{uuid "user"}}, {{uuid "cm7"}}, 0),
		({{uuid "user"}}, {{uuid "cm8"}}, 0),
		({{uuid "user"}}, {{uuid "cm9"}}, 0),
		({{uuid "user"}}, {{uuid "cm10"}}, 0),
		({{uuid "user"}}, {{uuid "cm11"}}, 0),
		({{uuid "user"}}, {{uuid "cm12"}}, 0),
		({{uuid "user"}}, {{uuid "cm13"}}, 0),
		({{uuid "user"}}, {{uuid "cm14"}}, 0);

	insert into escalation_policies (id, name) 
	values
		({{uuid "eid"}}, 'esc policy');
	insert into escalation_policy_steps (id, escalation_policy_id) 
	values
		({{uuid "esid"}}, {{uuid "eid"}});
	insert into escalation_policy_actions (escalation_policy_step_id, user_id) 
	values 
		({{uuid "esid"}}, {{uuid "user"}});

	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');

	insert into alerts (service_id, description) 
	values
		({{uuid "sid"}}, 'testing');

`
	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	d1 := h.Twilio(t).Device(h.Phone("1"))
	d2 := h.Twilio(t).Device(h.PhoneCC("+91", "1"))
	d3 := h.Twilio(t).Device(h.PhoneCC("+49", "1"))
	d4 := h.Twilio(t).Device(h.PhoneCC("+44", "1"))
	d5 := h.Twilio(t).Device(h.PhoneCC("+852", "1"))
	d6 := h.Twilio(t).Device(h.PhoneCC("+86", "1"))
	d7 := h.Twilio(t).Device(h.PhoneCC("+502", "1"))

	d1.ExpectSMS("testing")
	d1.ExpectVoice("testing")

	d2.ExpectSMS("testing")
	d2.ExpectVoice("testing")

	d3.ExpectSMS("testing")
	d3.ExpectVoice("testing")

	d4.ExpectSMS("testing")
	d4.ExpectVoice("testing")

	d5.ExpectSMS("testing")
	d5.ExpectVoice("testing")

	d6.ExpectSMS("testing")
	d6.ExpectVoice("testing")

	d7.ExpectSMS("testing")
	d7.ExpectVoice("testing")
}
