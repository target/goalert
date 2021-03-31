package oncall_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/oncall"
)

func cloneSlice(a []string) []string {
	if len(a) == 0 {
		return nil
	}
	s := make([]string, len(a))
	copy(s, a)
	return s
}

func TestUserCalculator(t *testing.T) {
	type result struct {
		Time  time.Time
		Value []string
	}
	var (
		start = time.Date(2000, 1, 2, 3, 4, 0, 0, time.UTC)
		end   = time.Date(2000, 1, 2, 3, 8, 0, 0, time.UTC)
	)
	check := func(desc string, expected []result, setup func(*oncall.UserCalculator)) {
		t.Run(desc, func(t *testing.T) {
			iter := oncall.NewTimeIterator(
				start,
				end,
				time.Minute,
			).NewUserCalculator()

			if setup != nil {
				setup(iter)
			}

			iter.Init()

			var count int
			for iter.Next() {
				i := count
				count++
				if count > len(expected) {
					t.Errorf("unexpected result: Value=%v, Time=%d", iter.ActiveUsers(), iter.Unix())
					continue
				}
				if len(expected[i].Value) > 0 || len(iter.ActiveUsers()) > 0 {
					assert.EqualValuesf(t, expected[i].Value, iter.ActiveUsers(), "result[%d].Value", i)
				}
				assert.Equalf(t, expected[i].Time.String(), time.Unix(iter.Unix(), 0).UTC().String(), "result[%d].Time", i)
			}
			if count != len(expected) {
				t.Errorf("got %d results; want %d", count, len(expected))
			}
		})
	}

	check("empty",
		[]result{
			{Time: start},
			{Time: end},
		},
		nil,
	)

	check("simple",
		[]result{
			{Time: start},
			{Time: time.Date(2000, 1, 2, 3, 5, 0, 0, time.UTC), Value: []string{"foo"}},
			{Time: time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC)},
			{Time: end},
		},
		func(iter *oncall.UserCalculator) {
			iter.SetSpan(time.Date(2000, 1, 2, 3, 5, 0, 0, time.UTC), time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC), "foo")
		},
	)

	check("at-start",
		[]result{
			{Time: time.Date(2000, 1, 2, 3, 3, 0, 0, time.UTC), Value: []string{"foo"}},
			{Time: time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC)},
			{Time: end},
		},
		func(iter *oncall.UserCalculator) {
			iter.SetSpan(time.Date(2000, 1, 2, 3, 3, 0, 0, time.UTC), time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC), "foo")
		},
	)

	check("multiple",
		[]result{
			{Time: start},
			{Time: time.Date(2000, 1, 2, 3, 5, 0, 0, time.UTC), Value: []string{"foo"}},
			{Time: time.Date(2000, 1, 2, 3, 6, 0, 0, time.UTC), Value: []string{"foo", "bar"}},
			{Time: time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC)},
			{Time: end},
		},
		func(iter *oncall.UserCalculator) {
			iter.SetSpan(time.Date(2000, 1, 2, 3, 5, 0, 0, time.UTC), time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC), "foo")
			iter.SetSpan(time.Date(2000, 1, 2, 3, 6, 0, 0, time.UTC), time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC), "bar")
		},
	)
}
