package validate

import "testing"

func TestLabelKey(t *testing.T) {
	check := func(valid bool, values ...string) {
		for _, val := range values {
			t.Run(val, func(t *testing.T) {
				err := LabelKey("", val)
				if valid && err != nil {
					t.Errorf("got %v; want nil", err)
				} else if !valid && err == nil {
					t.Errorf("got nil; want err")
				}
			})
		}
	}

	check(true,
		"foo/bar", "bin.baz/abc", "0123/abc", "foo-bar/abc",
	)
	check(false,
		"test", "Foo/bar", "-test/ok", "a/b", "foo/ok", "foo'/bar", " ", "", "foo /bar",
		"/", "//", "/foo/", "/foo/bar", "foo/",
	)
}
