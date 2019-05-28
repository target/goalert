package integrationkey

import (
	"testing"
)

func TestIntegrationKey_Normalize(t *testing.T) {
	test := func(valid bool, k IntegrationKey) {
		name := "valid"
		if !valid {
			name = "invalid"
		}
		t.Run(name, func(t *testing.T) {
			t.Logf("%+v", k)
			_, err := k.Normalize()
			if valid && err != nil {
				t.Errorf("got %v; want nil", err)
			} else if !valid && err == nil {
				t.Errorf("got nil err; want non-nil")
			}
		})
	}

	valid := []IntegrationKey{
		{Name: "SampleIntegrationKey", ServiceID: "e93facc0-4764-012d-7bfb-002500d5d1a6", Type: TypeGrafana},
	}
	invalid := []IntegrationKey{
		{},
	}
	for _, k := range valid {
		test(true, k)
	}
	for _, k := range invalid {
		test(false, k)
	}
}
