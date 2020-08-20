package oncall_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/oncall"
)

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	m := make(map[string]struct{}, len(a))
	for _, id := range a {
		m[id] = struct{}{}
	}
	for _, id := range b {
		_, ok := m[id]
		if !ok {
			return false
		}
	}
	return true
}

func TestUserCalculator(t *testing.T) {
	check := func(desc string, results [][]string, setup func(*oncall.UserCalculator)) {
		t.Run(desc, func(t *testing.T) {
			iter := oncall.NewTimeIterator(
				time.Date(2000, 1, 2, 3, 4, 0, 0, time.UTC),
				time.Date(2000, 1, 2, 3, 8, 0, 0, time.UTC),
				time.Minute,
			).NewUserCalculator()

			if setup != nil {
				setup(iter)
			}

			iter.Init()

			var last []string
			for i, exp := range results {
				if exp == nil {
					exp = []string{}
				}
				assert.Truef(t, iter.Next(), "Next() call #%d", i+1)
				assert.Equalf(t, !equalSlice(last, exp), iter.Changed(), "Changed() call #%d", i+1)
				act := iter.ActiveUsers()
				if act == nil {
					act = []string{}
				}
				assert.EqualValuesf(t, exp, act, "Active() call #%d", i+1)
				last = exp
			}

			assert.Falsef(t, iter.Next(), "Next() call #%d (last)", len(results)+1)
		})
	}
	check("empty", [][]string{nil, nil, nil, nil, nil}, nil)

	check("simple", [][]string{nil, {"foo"}, {"foo"}, nil, nil}, func(iter *oncall.UserCalculator) {
		iter.SetSpan(time.Date(2000, 1, 2, 3, 5, 0, 0, time.UTC), time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC), "foo")
	})

	check("at-start", [][]string{{"foo"}, {"foo"}, {"foo"}, nil, nil}, func(iter *oncall.UserCalculator) {
		iter.SetSpan(time.Date(2000, 1, 2, 3, 3, 0, 0, time.UTC), time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC), "foo")
	})

	check("multiple", [][]string{nil, {"foo"}, {"foo", "bar"}, nil, nil}, func(iter *oncall.UserCalculator) {
		iter.SetSpan(time.Date(2000, 1, 2, 3, 5, 0, 0, time.UTC), time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC), "foo")
		iter.SetSpan(time.Date(2000, 1, 2, 3, 6, 0, 0, time.UTC), time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC), "bar")
	})
}
