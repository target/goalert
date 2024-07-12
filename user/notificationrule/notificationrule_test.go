package notificationrule

import (
	"testing"

	"github.com/google/uuid"
)

func TestNotificationRule_Normalize(t *testing.T) {
	test := func(valid bool, nr NotificationRule) {
		name := "valid"
		if !valid {
			name = "invalid"
		}
		t.Run(name, func(t *testing.T) {
			t.Logf("%+v", nr)
			_, err := nr.Normalize(false)
			if valid && err != nil {
				t.Errorf("got %v; want nil", err)
			} else if !valid && err == nil {
				t.Errorf("got nil err; want non-nil")
			}
		})
	}

	valid := []NotificationRule{
		{DelayMinutes: 5, ContactMethodID: uuid.MustParse("ececacc0-4764-012d-7bfb-002500d5dece"), UserID: "bcefacc0-4764-012d-7bfb-002500d5decb"},
	}
	invalid := []NotificationRule{
		{},
	}
	for _, nr := range valid {
		test(true, nr)
	}
	for _, nr := range invalid {
		test(false, nr)
	}
}
