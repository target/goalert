package validate

import (
	"testing"
)

func TestSanitizeName(t *testing.T) {
	check := func(name, exp string) {
		t.Run(name, func(t *testing.T) {
			t.Logf("Name='%s'", name)
			res := SanitizeName(name)
			if res != exp {
				t.Errorf("got '%s'; want '%s'", res, exp)
			}
		})
	}
	check(" foo", "foo")
	check("okay \b", "okay")
	check("okay\n\nthen", "okay then")
	check("foo-bar", "foo-bar")
}

func TestName(t *testing.T) {
	check := func(name string, ok bool) {
		t.Run(name, func(t *testing.T) {
			t.Logf("Name='%s'", name)
			err := Name("", name)
			if err != nil && ok {
				t.Errorf("got %v; want nil", err)
			} else if err == nil && !ok {
				t.Errorf("got nil; want err")
			}
		})
	}

	valid := []string{
		"foo",
		"bar-Bin",
		"baz ok",
		"o'hello",
		"абаза бызшва (abaza bəzš˚a)",
		"𒀝𒅗𒁺𒌑 (Akkadû)",
		"客家話 [客家话]",
		"Xaat Kíl",
	}
	for _, n := range valid {
		check(n, true)
	}

	invalid := []string{
		"",
		" a",
		"a ",
		"test\b",
	}
	for _, n := range invalid {
		check(n, false)
	}
}
