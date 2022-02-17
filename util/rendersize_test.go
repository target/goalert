package util

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderSize(t *testing.T) {
	check := func(max int, input, output string) {
		t.Helper()
		result, err := RenderSize(max, input, func(n string) (string, error) {
			return strings.ReplaceAll(n, "&", "&amp;"), nil
		})
		assert.NoError(t, err)
		assert.Equal(t, output, result)
	}

	check(10, "foobarbaz", "foobarbaz")
	check(10, "foobarbazbin", "foobarbazb")
	check(10, "foo&&rbazbin", "foo&amp;")
	check(10, "foobarba&&&&", "foobarba")

	_, err := RenderSize(10, "foo", func(string) (string, error) {
		return "", fmt.Errorf("failed")
	})
	assert.Error(t, err)

	_, err = RenderSize(5, "foo", func(string) (string, error) {
		return "123456", nil
	})
	// can't fulfill the request
	assert.Error(t, err)
}

func TestRenderSizeN(t *testing.T) {
	check := func(max int, inputs []string, output string) {
		t.Helper()
		result, err := RenderSizeN(max, inputs, func(inputs []string) (string, error) {
			return strings.ReplaceAll(strings.Join(inputs, ""), "&", "&amp;"), nil
		})
		assert.NoError(t, err)
		assert.Equal(t, output, result)
	}

	check(10, []string{"foobarbaz"}, "foobarbaz")
	check(10, []string{"foo", "bar", "baz", "bin"}, "fbarbazbin")
	check(8, []string{"foo", "bar", "baz", "bin"}, "babazbin")
}
