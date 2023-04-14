package validate

import (
	"testing"
)

func FuzzPhone(f *testing.F) {
	numbers := []string{
		"+17633453456",
		"+919632040000",
		"+17734562190",
		"+916301210000",
		"+447480809090",
		"+61455518786",   // Australia
		"+498963648018",  // Germany
		"+85268355559",   // Hong Kong
		"+8618555196185", // China
		"+50223753964",   // Guatemala
		"+10633453456",
		"+15555555555",
		"+4474808090",
		"+611111111111",
		"+491515559510",
		"+85211111111",
	}
	for _, number := range numbers {
		f.Add(number)
	}

	f.Fuzz(func(t *testing.T, number string) {
		_ = Phone("Number", number)
	})
}

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
		"+61455518786",   // Australia
		"+498963648018",  // Germany
		"+85268355559",   // Hong Kong
		"+8618555196185", // China
		"+50223753964",   // Guatemala
	}
	for _, number := range valid {
		check(number, true)
	}

	invalid := []string{
		"+10633453456",
		"+15555555555",
		"+4474808090",
		"+611111111111",
		"+491515559510",
		"+85211111111",
	}
	for _, number := range invalid {
		check(number, false)
	}
}
