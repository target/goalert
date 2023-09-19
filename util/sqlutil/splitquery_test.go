package sqlutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitQuery(t *testing.T) {
	assert.Equal(t, []string{"foobar"}, SplitQuery("foobar"))
	assert.Equal(t, []string{"foo", "bar"}, SplitQuery("foo;bar"))
	assert.Equal(t, []string{"foo", "bar"}, SplitQuery("foo;bar;"))
	assert.Equal(t, []string{"foo$$; $$bar", "baz"}, SplitQuery("foo$$; $$bar;baz"))
}
