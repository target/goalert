package escalation

import (
	"testing"
)

func TestStep_Normalize(t *testing.T) {
	test := func(valid bool, s Step) {
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

	valid := []Step{
		{PolicyID: "a81facc0-4764-012d-7bfb-002500d5d678", DelayMinutes: 1},
	}

	invalid := []Step{
		{PolicyID: "a81facc0-4764-012d-7bfb-002500d5d678", DelayMinutes: 9001},
	}
	for _, s := range valid {
		test(true, s)
	}
	for _, s := range invalid {
		test(false, s)
	}
}
