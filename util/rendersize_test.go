package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderSize(t *testing.T) {
	var input string
	renderFunc := func(n int) string {
		if n > len(input) {
			n = len(input)
		}
		return strings.ReplaceAll(input[:n], "&", "&amp;")
	}

	input = "foobarbaz"
	result := RenderSize(10, renderFunc)
	assert.Equal(t, "foobarbaz", result)

	input = "foobarbazbin"
	result = RenderSize(10, renderFunc)
	assert.Equal(t, "foobarbazb", result)

	input = "foo&&rbazbin"
	result = RenderSize(10, renderFunc)
	assert.Equal(t, "foo&amp;", result)

	input = "foobarba&&&&"
	result = RenderSize(10, renderFunc)
	assert.Equal(t, "foobarba", result)
}
