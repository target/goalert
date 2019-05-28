package schedule

import (
	"testing"
	"time"
)

func TestSchedule_Normalize(t *testing.T) {
	test := func(valid bool, name string, s Schedule) {
		t.Run(name, func(t *testing.T) {
			_, err := s.Normalize()
			if valid && err != nil {
				t.Errorf("err = %v; want nil", err)
			} else if !valid && err == nil {
				t.Errorf("err = nil; want != nil")
			}
		})
	}

	data := []struct {
		v bool
		n string
		s Schedule
	}{
		{false, "missing name", Schedule{Description: "hello", TimeZone: time.Local}},
	}
	for _, d := range data {
		test(d.v, d.n, d.s)
	}
}
