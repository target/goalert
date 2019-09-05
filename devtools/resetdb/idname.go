package main

import (
	"fmt"

	"github.com/brianvoe/gofakeit"
	"github.com/target/goalert/validation/validate"
)

// idName will return a function that will generate a random name/ID for something like a rotation or schedule with a suffix.
func idName(suffix string) func() string {
	return func() string {
		var res string
		for {
			res = fmt.Sprintf("%s %s %s %s", gofakeit.JobDescriptor(), gofakeit.BuzzWord(), gofakeit.JobLevel(), suffix)
			err := validate.IDName("", res)
			if err == nil {
				return res
			}
		}
	}
}
