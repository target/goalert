package util

import (
	"testing"
)

func TestJoinURL(t *testing.T) {
	test := func(name, base, suffix, expected string) {
		t.Run(name, func(t *testing.T) {
			result, err := JoinURL(base, suffix)
			if err != nil {
				t.Fatalf("err = %v; want nil", err)
			}

			if result != expected {
				t.Errorf("result = '%s'; want '%s'", result, expected)
			}
		})
	}

	test("no trailing slash", "http://foo.bar", "/baz", "http://foo.bar/baz")
	test("both slashes", "http://foo.bar/", "/baz", "http://foo.bar/baz")
	test("no slashes", "http://foo.bar", "baz", "http://foo.bar/baz")
	test("trailing slash", "http://foo.bar/", "baz", "http://foo.bar/baz")
	test("query param on base", "http://foo.bar?a=b", "/baz", "http://foo.bar/baz?a=b")
	test("query param on suffix", "http://foo.bar", "/baz?c=d", "http://foo.bar/baz?c=d")
	test("both query params", "http://foo.bar?a=b", "/baz?c=d", "http://foo.bar/baz?a=b&c=d")
}
