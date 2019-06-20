package validate

import (
	"testing"
)

func TestPhone(t *testing.T) {
	check := func(number string, expValid bool) {
		name := "valid"
		if !expValid {
			name = "invalid"
		}
		t.Run(name, func(t *testing.T) {
			err := Phone("", number)
			if expValid && err != nil {
				t.Errorf("got %v; want %s to be valid (nil err)", err, number)
			} else if !expValid && err == nil {
				t.Errorf("got nil; want %s to be invalid", number)
			}
		})
	}

	valid := []string{
		"+17633453456",
		"+919632040000",
		"+17734562190",
		"+916301210000",
		"+447480809090",
	}
	for _, number := range valid {
		check(number, true)
	}

	invalid := []string{
		"+10633453456",
		"+15555555555",
		"+4474808090",
	}
	for _, number := range invalid {
		check(number, false)
	}

}
