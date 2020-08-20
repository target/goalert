package oncall_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/override"
)

func TestOverrideCalculator(t *testing.T) {
	later := time.Date(9999, 1, 0, 0, 0, 0, 0, time.UTC)
	overrides := []override.UserOverride{
		{AddUserID: "foo", End: later},
		{AddUserID: "bar", RemoveUserID: "rab", End: later},
		{RemoveUserID: "baz", End: later},
	}
	iter := oncall.NewTimeIterator(
		time.Date(2000, 1, 2, 3, 4, 0, 0, time.UTC),
		time.Date(2000, 1, 2, 3, 8, 0, 0, time.UTC),
		time.Minute,
	).NewOverrideCalculator(overrides)

	assert.True(t, iter.Next(), "Next()")

	result := iter.MapUsers([]string{"rab", "baz", "bin"})
	assert.EqualValues(t, []string{"bar", "bin", "foo"}, result)
}
