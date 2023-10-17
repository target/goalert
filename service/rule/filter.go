package rule

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/ast"
	"github.com/antonmedv/expr/parser"
	"github.com/pkg/errors"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

type Filter struct {
	Field     string
	Operator  string
	Value     string
	ValueType FilterValueType
}

const (
	filterDelimeter = " && "
)

// FiltersToExprString takes in a slice of filters, validates them, and returns the
// valid expr string that represents them
func FiltersToExprString(filters []Filter) (string, error) {
	// no filters means the expression should always evalutate to true
	if len(filters) == 0 {
		return "true", nil
	}

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
			continue
		}

		// ensure filter string compiles
		if _, err := expr.Compile(fStr); err != nil {
			errs = append(errs, validation.NewFieldError(fmt.Sprintf("Filters[%d]", i), err.Error()))
			continue
		}

		str += fStr
	}

	err := validate.Many(errs...)
	if err != nil {
		return "", err
	}

	return str, nil
}

func filterToExprString(filter Filter) (string, error) {
	// slice of errors lets us validate field/operator/value all at once
	errs := make([]error, 3)
	errs[0] = validate.FilterField("Field", filter.Field)

	value := filter.Value
	switch filter.ValueType {
	case TypeString:
		value = quoteAndEscapeValue(value)
		errs[1] = validate.FilterStringOperator("Operator", filter.Operator)
	case TypeNumber:
		errs[1] = validate.FilterNumberOperator("Operator", filter.Operator)
		errs[2] = validate.FilterNumberValue("Value", value)
	case TypeBool:
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

// FiltersFromExprString returns a slice of Filter structs parsed from the
// given expr string (which should be ' && ' delimited)
func FiltersFromExprString(str string) ([]Filter, error) {
	// always true expression means there are no filters
	if str == "true" {
		return []Filter{}, nil
	}

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

type FilterValueType int

const (
	TypeUnknown FilterValueType = iota
	TypeString
	TypeNumber
	TypeBool
)

// UnmarshalGQL implements the graphql.Marshaler interface
func (t *FilterValueType) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err != nil {
		return err
	}

	switch str {
	case "UNKNOWN":
		*t = TypeUnknown
	case "STRING":
		*t = TypeString
	case "NUMBER":
		*t = TypeNumber
	case "BOOL":
		*t = TypeBool
	default:
		return validation.NewFieldError("Type", "unknown type "+str)
	}

	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (t FilterValueType) MarshalGQL(w io.Writer) {
	switch t {
	case TypeUnknown:
		graphql.MarshalString("UNKNOWN").MarshalGQL(w)
	case TypeString:
		graphql.MarshalString("STRING").MarshalGQL(w)
	case TypeNumber:
		graphql.MarshalString("NUMBER").MarshalGQL(w)
	case TypeBool:
		graphql.MarshalString("BOOL").MarshalGQL(w)
	}
}
