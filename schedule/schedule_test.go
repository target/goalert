package schedule

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSearchOptions_Only(t *testing.T) {
	validUUID := "A035FD3C-73C8-4F72-BECD-36B027AE1374"

	_, err := (*renderData)(&SearchOptions{Only: []string{validUUID}}).Normalize()
	require.NoError(t, err)

	_, err = (*renderData)(&SearchOptions{Only: []string{"not-a-uuid"}}).Normalize()
	require.Error(t, err)

	tooMany := make([]string, 51)
	for i := range tooMany {
		tooMany[i] = validUUID
	}
	_, err = (*renderData)(&SearchOptions{Only: tooMany}).Normalize()
	require.Error(t, err)

	args := (*renderData)(&SearchOptions{Only: []string{validUUID}}).QueryArgs()
	require.Condition(t, func() bool {
		for _, a := range args {
			if a.Name == "only" {
				return a.Value != nil
			}
		}
		return false
	}, "QueryArgs() should include a non-nil 'only' parameter")
}

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
