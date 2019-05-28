package validate

import (
	"testing"
)

func TestIDName(t *testing.T) {
	test := func(valid bool, n string) {
		var title string
		if valid {
			title = "Valid"
		} else {
			title = "Invalid"
		}
		t.Run(title, func(t *testing.T) {
			err := IDName("Name", n)
			if err == nil && !valid {
				t.Errorf("IDName(%s) = nil; want err", n)
			} else if err != nil && valid {
				t.Errorf("IDName(%s) = %v; want nil", n, err)
			}
		})
	}

	invalid := []string{
		"", "  ", " 5", "5for", "_asdf", ",asdf", "asdf\nasdf", "end ",
		"a#%$^#$%&#2a",
	}
	valid := []string{
		"a-_' 0",
	}
	for _, n := range invalid {
		test(false, n)
	}
	for _, n := range valid {
		test(true, n)
	}
}
