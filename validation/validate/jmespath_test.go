package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJMESPath(t *testing.T) {
	err := JMESPath("test", "foobar")
	assert.NoError(t, err)

	err = JMESPath("test", "foo.bar")
	assert.NoError(t, err)

	err = JMESPath("test", "foobar || ")
	assert.Error(t, err)

}
