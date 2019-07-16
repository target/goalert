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
	validIndia := []ContactMethod{
		{Name: "validIndia", Type: TypeSMS, Value: "+918105554545"},
	}
	invalidIndia := []ContactMethod{
		{Name: "validIndia", Type: TypeSMS, Value: "+918105554545a"},
	}
	validUK := []ContactMethod{
		{Name: "validUK", Type: TypeSMS, Value: "+447911123456"},
	}
	invalidUK := []ContactMethod{
		{Name: "validUK", Type: TypeSMS, Value: "+448105554545"},
	}

	for _, cm := range valid {
		test(true, cm)
	}
	for _, cm := range invalid {
		test(false, cm)
	}
	for _, cm := range validIndia {
		test(true, cm)
	}
	for _, cm := range invalidIndia {
		test(false, cm)
	}
	for _, cm := range validUK {
		test(true, cm)
	}
	for _, cm := range invalidUK {
		test(false, cm)
	}
}
