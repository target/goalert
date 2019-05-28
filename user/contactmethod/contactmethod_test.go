package contactmethod

import (
	"testing"
)

func TestContactMethod_Normalize(t *testing.T) {
	test := func(valid bool, cm ContactMethod) {
		name := "valid"
		if !valid {
			name = "invalid"
		}
		t.Run(name, func(t *testing.T) {
			t.Logf("%+v", cm)
			_, err := cm.Normalize()
			if valid && err != nil {
				t.Errorf("got %v; want nil", err)
			} else if !valid && err == nil {
				t.Errorf("got nil err; want non-nil")
			}
		})
	}

	valid := []ContactMethod{
		{Name: "Iphone", Type: TypeSMS, Value: "+15515108117"},
	}
	invalid := []ContactMethod{
		{Name: "abcd", Type: TypeSMS, Value: "+15555555555"},
	}
	for _, cm := range valid {
		test(true, cm)
	}
	for _, cm := range invalid {
		test(false, cm)
	}
}
