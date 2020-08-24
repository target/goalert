package oncall_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/oncall"
)

func TestActiveCalculator(t *testing.T) {
	check := func(desc string, results []bool, setup func(*oncall.ActiveCalculator)) {
		t.Helper()
		t.Run(desc, func(t *testing.T) {
			t.Helper()
			iter := oncall.NewTimeIterator(
				time.Date(2000, 1, 2, 3, 4, 0, 0, time.UTC),
				time.Date(2000, 1, 2, 3, 8, 0, 0, time.UTC),
				time.Minute,
			).NewActiveCalculator()

			if setup != nil {
				setup(iter)
			}
			iter.Init()

			var last bool
			for i, exp := range results {
				assert.Truef(t, iter.Next(), "Next() call #%d", i+1)
				assert.Equalf(t, exp != last, iter.Changed(), "Changed() call #%d", i+1)
				assert.Equalf(t, exp, iter.Active(), "Active() call #%d", i+1)
				last = exp
			}

			assert.Falsef(t, iter.Next(), "Next() call #%d (last)", len(results)+1)
		})
	}
	check("empty", []bool{false, false, false, false, false}, nil)

	check("simple", []bool{false, true, true, false, false}, func(iter *oncall.ActiveCalculator) {
		iter.SetSpan(time.Date(2000, 1, 2, 3, 5, 0, 0, time.UTC), time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC))
	})

	check("at-start", []bool{true, true, true, false, false}, func(iter *oncall.ActiveCalculator) {
		iter.SetSpan(time.Date(2000, 1, 2, 3, 3, 0, 0, time.UTC), time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC))
	})

	check("multiple", []bool{false, true, false, true, false}, func(iter *oncall.ActiveCalculator) {
		iter.SetSpan(time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC), time.Date(2000, 1, 2, 3, 8, 0, 0, time.UTC))
		// out of order
		iter.SetSpan(time.Date(2000, 1, 2, 3, 5, 0, 0, time.UTC), time.Date(2000, 1, 2, 3, 6, 0, 0, time.UTC))
	})

	check("full", []bool{true, true, true, true, true}, func(iter *oncall.ActiveCalculator) {
		iter.SetSpan(time.Time{}, time.Date(9999, 1, 0, 0, 0, 0, 0, time.UTC))
	})

	t.Run("ActiveTime", func(t *testing.T) {
		iter := oncall.NewTimeIterator(
			time.Date(2000, 1, 2, 3, 4, 0, 0, time.UTC),
			time.Date(2000, 1, 2, 3, 8, 0, 0, time.UTC),
			time.Minute,
		).NewActiveCalculator()

		iter.SetSpan(time.Date(2000, 1, 2, 3, 1, 0, 0, time.UTC), time.Date(2000, 1, 2, 3, 6, 0, 0, time.UTC))

		iter.Init()

		assert.True(t, iter.Next())
		assert.True(t, iter.Changed())
		assert.True(t, iter.Active())
		assert.Equal(t, time.Date(2000, 1, 2, 3, 1, 0, 0, time.UTC).Unix(), iter.ActiveTime().Unix())

	})

}
