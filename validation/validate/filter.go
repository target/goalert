package validate

import (
	"regexp"
	"strconv"

	"github.com/target/goalert/validation"
)

var (
	fieldPattern = regexp.MustCompile(`^[^\t\n\\&"]+$`)
)

func FilterField(fname, field string) error {
	if !fieldPattern.MatchString(field) {
		return validation.NewFieldError(fname, "field contains invalid characters")
	}
	return nil
}

func FilterStringOperator(fname, operator string) error {
	switch operator {
	case "==", "!=", "contains", "startsWith", "endsWith":
		return nil
	default:
		return validation.NewFieldError(fname, "bad operator for string value")
	}
}

func FilterNumberOperator(fname, operator string) error {
	switch operator {
	case "==", "!=", "<", "<=", ">", ">=":
		return nil
	default:
		return validation.NewFieldError(fname, "bad operator for number value")
	}
}

func FilterNumberValue(fname, value string) error {
	_, err := strconv.ParseFloat(value, 64)
	if err == nil {
		return nil
	}
	_, err = strconv.ParseInt(value, 10, 64)
	if err == nil {
		return nil
	}
	return validation.NewFieldError(fname, "invalid number value")
}

func FilterBoolOperator(fname, operator string) error {
	switch operator {
	case "==", "!=":
		return nil
	default:
		return validation.NewFieldError(fname, "bad operator for boolean value")
	}
}

func FilterBoolValue(fname, value string) error {
	if value != "true" && value != "false" {
		return validation.NewFieldError(fname, "invalid boolean value")
	}
	return nil
}
