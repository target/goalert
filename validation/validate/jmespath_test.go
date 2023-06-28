package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func FuzzJMESPath(f *testing.F) {
	invalid := []string{
		"foobar || ", "foo.bar", "foobar",
	}
	for _, s := range invalid {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		_ = JMESPath("Name", s)
	})
}

func TestJMESPath(t *testing.T) {
	err := JMESPath("test", "foobar")
	assert.NoError(t, err)

	err = JMESPath("test", "foo.bar")
	assert.NoError(t, err)

	err = JMESPath("test", "foobar || ")
	assert.Error(t, err)
}
