package validate

import (
	"unicode"

	"github.com/jmespath/go-jmespath"
	"github.com/target/goalert/validation"
)

// JMESPath will validate a JMESPath expression.
func JMESPath(fname, expression string) error {
	for _, c := range expression {
		if !unicode.IsPrint(c) && c != '\t' && c != '\n' {
			return validation.NewFieldError(fname, "only printable characters allowed")
		}
	}

	_, err := jmespath.Compile(expression)
	if err != nil {
		return validation.NewFieldError(fname, err.Error())
	}

	return nil
}
