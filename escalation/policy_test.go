package escalation

import (
	"testing"
)

func TestPolicy_Normalize(t *testing.T) {
	test := func(valid bool, p Policy) {
		name := "valid"
		if !valid {
			name = "invalid"
		}
		t.Run(name, func(t *testing.T) {
			t.Logf("%+v", p)
			_, err := p.Normalize()
			if valid && err != nil {
				t.Errorf("got %v; want nil", err)
			} else if !valid && err == nil {
				t.Errorf("got nil err; want non-nil")
			}
		})
	}

	valid := []Policy{
		{Name: "SampleEscPolicy", Description: "Sample Escalation Policy", Repeat: 1},
	}
	invalid := []Policy{
		{Name: "SampleEscPolicy", Description: "Sample Escalation Policy", Repeat: -5},
	}
	for _, p := range valid {
		test(true, p)
	}
	for _, p := range invalid {
		test(false, p)
	}
}
