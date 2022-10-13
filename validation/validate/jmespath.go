package validate

import (
	"github.com/jmespath/go-jmespath"
	"github.com/target/goalert/validation"
)

// JMESPath will validate a JMESPath expression.
func JMESPath(fname, expression string) error {
	err := Text(fname, expression, 0, 4096)
	if err != nil {
		return err
	}

	_, err = jmespath.Compile(expression)
	if err != nil {
		return validation.NewFieldError(fname, err.Error())
	}

	return nil
}
