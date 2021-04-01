package oncall_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/oncall"
)

func TestActiveCalculator(t *testing.T) {
	var (
		start = time.Date(2000, 1, 2, 3, 4, 0, 0, time.UTC)
		end   = time.Date(2000, 1, 2, 3, 8, 0, 0, time.UTC)
	)
	type result struct {
		Time  time.Time
		Value bool
	}

	check := func(desc string, expected []result, setup func(*oncall.ActiveCalculator)) {
		t.Helper()
		t.Run(desc, func(t *testing.T) {
			t.Helper()
			iter := oncall.NewTimeIterator(
				start,
				end,
				time.Minute,
			).NewActiveCalculator()

			if setup != nil {
				setup(iter)
			}
			iter.Init()

			var count int
			for iter.Next() {
				i := count
				count++
				if count > len(expected) {
					t.Errorf("unexpected result: Value=%t, Time=%d", iter.Active(), iter.Unix())
					continue
				}
				assert.Equalf(t, expected[i].Value, iter.Active(), "result[%d].Value", i)
				assert.Equalf(t, expected[i].Time.String(), time.Unix(iter.Unix(), 0).UTC().String(), "result[%d].Time", i)
			}
			if count != len(expected) {
				t.Errorf("got %d results; want %d", count, len(expected))
			}
		})
	}
	check("empty", []result{{Time: start}, {Time: end}}, nil)

	check("simple",
		[]result{
			{Time: start},
			{Time: time.Date(2000, 1, 2, 3, 5, 0, 0, time.UTC), Value: true},
			{Time: time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC), Value: false},
			{Time: end},
		},
		func(iter *oncall.ActiveCalculator) {
			iter.SetSpan(time.Date(2000, 1, 2, 3, 5, 0, 0, time.UTC), time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC))
		},
	)

	check("at-start",
		[]result{
			{Time: time.Date(2000, 1, 2, 3, 3, 0, 0, time.UTC), Value: true},
			{Time: time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC)},
			{Time: end},
		},
		func(iter *oncall.ActiveCalculator) {
			iter.SetSpan(time.Date(2000, 1, 2, 3, 3, 0, 0, time.UTC), time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC))
		},
	)

	check("multiple",
		[]result{
			{Time: start},
			{Time: time.Date(2000, 1, 2, 3, 5, 0, 0, time.UTC), Value: true},
			{Time: time.Date(2000, 1, 2, 3, 6, 0, 0, time.UTC), Value: false},
			{Time: time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC), Value: true},
			{Time: end},
		},
		func(iter *oncall.ActiveCalculator) {
			iter.SetSpan(time.Date(2000, 1, 2, 3, 5, 0, 0, time.UTC), time.Date(2000, 1, 2, 3, 6, 0, 0, time.UTC))
			iter.SetSpan(time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC), time.Date(2000, 1, 2, 3, 8, 0, 0, time.UTC))
		},
	)

	check("full",
		[]result{
			{Value: true},
			{Time: end, Value: true},
		},
		func(iter *oncall.ActiveCalculator) {
			iter.SetSpan(time.Time{}, time.Date(9999, 1, 0, 0, 0, 0, 0, time.UTC))
		},
	)

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

	})

}
