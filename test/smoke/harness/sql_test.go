package harness

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQLSplit(t *testing.T) {
	assert.Equal(t, []string{"foobar"}, sqlSplit("foobar"))
	assert.Equal(t, []string{"foo", "bar"}, sqlSplit("foo;bar"))
	assert.Equal(t, []string{"foo$$; $$bar", "baz"}, sqlSplit("foo$$; $$bar;baz"))
}
