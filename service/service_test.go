package service

import (
	"testing"
)

func TestService_Normalize(t *testing.T) {
	test := func(valid bool, s Service) {
		name := "valid"
		if !valid {
			name = "invalid"
		}
		t.Run(name, func(t *testing.T) {
			t.Logf("%+v", s)
			_, err := s.Normalize()
			if valid && err != nil {
				t.Errorf("got %v; want nil", err)
			} else if !valid && err == nil {
				t.Errorf("got nil err; want non-nil")
			}
		})
	}

	valid := []Service{
		{Name: "Sample Service", Description: "Sample Service", EscalationPolicyID: "A035FD3C-73C8-4F72-BECD-36B027AE1374"},
	}
	invalid := []Service{
		{},
	}
	for _, s := range valid {
		test(true, s)
	}
	for _, s := range invalid {
		test(false, s)
	}
}
