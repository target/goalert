package alert

import (
	"testing"
)

func TestAlert_Normalize(t *testing.T) {
	test := func(valid bool, a Alert) {
		name := "valid"
		if !valid {
			name = "invalid"
		}
		t.Run(name, func(t *testing.T) {
			t.Logf("%+v", a)
			_, err := a.Normalize()
			if valid && err != nil {
				t.Errorf("got %v; want nil", err)
			} else if !valid && err == nil {

				t.Errorf("got nil err; want non-nil")
			}
		})
	}

	valid := []Alert{
		{Summary: "Sample First Alert", Source: SourceManual, Status: StatusTriggered, ServiceID: "e93facc0-4764-012d-7bfb-002500d5d1a6"},
	}
	invalid := []Alert{
		{ServiceID: "e93facc0-4764-012d-7bfb"},
	}
	for _, a := range valid {
		test(true, a)
	}
	for _, a := range invalid {
		test(false, a)
	}
}
