package jsonutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApply(t *testing.T) {
	const rawDoc = `{"Bin": "baz","Foo": "bar"}`
	var newDoc struct {
		Foo string
	}
	newDoc.Foo = "FOO"

	data, err := Apply([]byte(rawDoc), newDoc)
	assert.NoError(t, err)
	assert.Equal(t, `{"Bin":"baz","Foo":"FOO"}`, string(data))

	data, err = Apply(nil, newDoc)
	assert.NoError(t, err)
	assert.Equal(t, `{"Foo":"FOO"}`, string(data))
}
