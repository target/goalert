package mocktwilio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRelURL(t *testing.T) {
	check := func(oldURL, newURL, expected string) {
		t.Helper()
		assert.Equal(t, expected, relURL(oldURL, newURL))
	}

	// absolute
	check("http://example.com/foo/bar", "http://other.example.org/bin/baz", "http://other.example.org/bin/baz")

	// relative
	check("http://example.com/foo/bar", "baz", "http://example.com/foo/baz")
	check("http://example.com/foo/bar", "../baz", "http://example.com/baz")
	check("http://example.com/foo/bar/", "baz", "http://example.com/foo/bar/baz")
	check("http://example.com/foo/bar/", "../baz", "http://example.com/foo/baz")

	// absolute path
	check("http://example.com/foo/bar/", "/bin", "http://example.com/bin")

	// invalid
	check("http://example.com/foo/bar/", "/../bin", "")
	check("http://example.com/foo/bar", "../../../../baz", "")
}
