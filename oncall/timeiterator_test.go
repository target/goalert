package oncall_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/oncall"
)

func TestTimeIterator_Step(t *testing.T) {
	iter := oncall.NewTimeIterator(time.Time{}, time.Time{}, time.Minute)
	assert.Equal(t, time.Minute, iter.Step())
}

func TestTimeIterator_Start(t *testing.T) {
	iter := oncall.NewTimeIterator(time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC), time.Time{}, time.Minute)
	assert.Equal(t, time.Date(2000, 1, 2, 3, 4, 0, 0, time.UTC).Unix(), iter.Start().Unix())
}
func TestTimeIterator_End(t *testing.T) {
	iter := oncall.NewTimeIterator(time.Time{}, time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC), time.Minute)
	assert.Equal(t, time.Date(2000, 1, 2, 3, 4, 0, 0, time.UTC).Unix(), iter.End().Unix())
}

func TestTimeIterator_Unix(t *testing.T) {
	iter := oncall.NewTimeIterator(time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC), time.Date(2001, 1, 2, 3, 4, 5, 6, time.UTC), time.Minute)
	iter.Next()

	assert.Equal(t, time.Date(2000, 1, 2, 3, 4, 0, 0, time.UTC).Unix(), iter.Unix())
}

func TestTimeIterator_Next(t *testing.T) {
	iter := oncall.NewTimeIterator(time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC), time.Date(2000, 1, 2, 3, 8, 5, 6, time.UTC), time.Minute)

	assert.True(t, iter.Next())
	assert.Equal(t, time.Date(2000, 1, 2, 3, 4, 0, 0, time.UTC).Unix(), iter.Unix())
	assert.True(t, iter.Next())
	assert.Equal(t, time.Date(2000, 1, 2, 3, 5, 0, 0, time.UTC).Unix(), iter.Unix())
	assert.True(t, iter.Next())
	assert.Equal(t, time.Date(2000, 1, 2, 3, 6, 0, 0, time.UTC).Unix(), iter.Unix())
	assert.True(t, iter.Next())
	assert.Equal(t, time.Date(2000, 1, 2, 3, 7, 0, 0, time.UTC).Unix(), iter.Unix())
	assert.True(t, iter.Next())
	assert.Equal(t, time.Date(2000, 1, 2, 3, 8, 0, 0, time.UTC).Unix(), iter.Unix())
	assert.False(t, iter.Next())
}
func TestTimeIterator_Register(t *testing.T) {
	iter := oncall.NewTimeIterator(time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC), time.Date(2000, 1, 2, 3, 8, 5, 6, time.UTC), time.Minute)

	var called bool
	iter.Register(func(unix int64) {
		assert.Equal(t, time.Date(2000, 1, 2, 3, 4, 0, 0, time.UTC).Unix(), unix)
		called = true
	}, nil)

	iter.Next()
	assert.True(t, called)
}
