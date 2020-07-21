package validate

import (
	"github.com/jmespath/go-jmespath"
	"github.com/target/goalert/validation"
)

// JMESPath will validate a JMESPath expression.
func JMESPath(fname, expression string) error {
	if expression == "" {
		return nil
	}

	_, err := jmespath.Compile(expression)
	if err != nil {
		return validation.NewFieldError(fname, err.Error())
	}

	return nil
}
