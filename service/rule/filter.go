package rule

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/ast"
	"github.com/antonmedv/expr/parser"
	"github.com/pkg/errors"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

type ValueType int

type Filter struct {
	Field     string
	Operator  string
	Value     string
	ValueType ValueType
}

const (
	UnknownType ValueType = iota
	StringType
	NumberType
	BoolType
)

const (
	filterDelimeter = " && "
)

// ToExprString takes in a slice of filters, validates them, and returns the
// valid expr string that represents them
func ToExprString(filters []Filter) (string, error) {
	str := ""

	// slice of errors lets us validate all filters at once
	errs := make([]error, 0, len(filters))
	for i, filter := range filters {
		if i > 0 {
			str += filterDelimeter
		}

		fStr, err := filterToExprString(filter)
		if err != nil {
			errs = append(errs, validation.AddPrefix(fmt.Sprintf("Filters[%d].", i), err))
		}

		str += fStr
	}

	err := validate.Many(errs...)
	if err != nil {
		return "", err
	}

	if _, err := expr.Compile(str); err != nil {
		log.Log(context.Background(), errors.Wrapf(err, "compile filter expr '%s'", str))
		return "", fmt.Errorf("compile filters expr")
	}
	return str, nil
}

func filterToExprString(filter Filter) (string, error) {
	// slice of errors lets us validate field/operator/value all at once
	errs := make([]error, 3)
	errs[0] = validate.FilterField("Field", filter.Field)

	value := filter.Value
	switch filter.ValueType {
	case StringType:
		value = quoteAndEscapeValue(value)
		errs[1] = validate.FilterStringOperator("Operator", filter.Operator)
	case NumberType:
		errs[1] = validate.FilterNumberOperator("Operator", filter.Operator)
		errs[2] = validate.FilterNumberValue("Value", value)
	case BoolType:
		value = strings.ToLower(filter.Value)
		errs[1] = validate.FilterBoolOperator("Operator", filter.Operator)
		errs[2] = validate.FilterBoolValue("Value", value)
	default:
		errs[1] = validation.NewFieldError("ValueType", "invalid value type")
	}

	if err := validate.Many(errs...); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s %s", filter.Field, filter.Operator, value), nil
}

// FromExprString
func FromExprString(str string) ([]Filter, error) {
	tree, err := parser.Parse(str)
	if err != nil {
		return nil, errors.Wrap(err, "parse expr tree")
	}

	// using expr's own syntax tree parsing package (ast) with our own custom
	// Visitor (FilterParser)
	filters := []Filter{}
	ast.Walk(&tree.Node, FilterParser{Filters: &filters, Err: &err})
	if err != nil {
		return nil, errors.Wrap(err, "get filters from tree")
	}

	return filters, err
}

func quoteAndEscapeValue(value string) string {
	quoted := strconv.Quote(value)
	// when parsing, strconv.Unquote will replace '\u0026' with '&' automatically
	return strings.ReplaceAll(quoted, "&", `\u0026`)
}
