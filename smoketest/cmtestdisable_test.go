package smoketest

import (
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

// TestCMTestDisable checks for a condition where a test message is scheduled but the contact method is immediately disabled.
func TestCMTestDisable(t *testing.T) {
	t.Parallel()

	sqlQuery := `
		insert into users (id, name, email) 
		values 
			({{uuid "user"}}, 'bob', 'joe');
		insert into user_contact_methods (id, user_id, name, type, value, disabled) 
		values
			({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}}, true),
			({{uuid "cm2"}}, {{uuid "user"}}, 'personal2', 'SMS', {{phone "2"}}, false);
		insert into outgoing_messages (message_type, contact_method_id, last_status, user_id)
		values
			('test_notification', {{uuid "cm1"}}, 'pending', {{uuid "user"}}),
			('test_notification', {{uuid "cm2"}}, 'pending', {{uuid "user"}});
	`
	h := harness.NewHarness(t, sqlQuery, "add-verification-code")
	defer h.Close()

	h.Twilio(t).Device(h.Phone("2")).ExpectSMS("test")
}
