package contactmethod

import (
	"context"
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
			_, err := cm.Normalize(context.Background(), nil)
			if valid && err != nil {
				t.Errorf("got %v; want nil", err)
			} else if !valid && err == nil {
				t.Errorf("got nil err; want non-nil")
			}
		})
	}

	// TODO: add validation for TypeEmail
	valid := []ContactMethod{
		{Name: "Iphone", Type: TypeSMS, Value: "+15515108117"},
		{Name: "validIndia", Type: TypeSMS, Value: "+918105554545"},
		{Name: "validUK", Type: TypeSMS, Value: "+447911123456"},

		{Name: "webhookHTTP", Type: TypeWebhook, Value: "http://www.example.com"},
		{Name: "webhookHTTPS", Type: TypeWebhook, Value: "https://www.example.com"},
		{Name: "webhookPath", Type: TypeWebhook, Value: "http://www.example.com/example"},
	}
	invalid := []ContactMethod{
		{Name: "abcd", Type: TypeSMS, Value: "+15555555555"},
		{Name: "invalidIndia", Type: TypeSMS, Value: "+918105554545a"},
		{Name: "invalidUK", Type: TypeSMS, Value: "+448105554545"},

		{Name: "webhookEmpty", Type: TypeWebhook, Value: ""},
		{Name: "webhookIncomplete", Type: TypeWebhook, Value: "example"},
		{Name: "webhookMissingProtocol", Type: TypeWebhook, Value: "example.com"},
	}

	for _, cm := range valid {
		test(true, cm)
	}
	for _, cm := range invalid {
		test(false, cm)
	}
}
