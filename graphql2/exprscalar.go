package graphql2

import (
	"reflect"

	"github.com/99designs/gqlgen/graphql"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
	"github.com/target/goalert/validation"
)

// ExprIsID returns true if the node is a valid identifier.
func ExprIsID(n ast.Node) bool {
	switch t := n.(type) {
	case *ast.IdentifierNode:
		return true
	case *ast.MemberNode:
		return !t.Method && ExprIsID(t.Node) && (ExprIsID(t.Property) || ExprIsLiteral(t.Property))
	case *ast.ChainNode:
		return ExprIsID(t.Node)
	}

	return false
}

// ExprIsLiteral returns true if the node is a literal value (scalar or array of scalar values).
func ExprIsLiteral(n ast.Node) bool {
	switch t := n.(type) {
	case *ast.StringNode:
	case *ast.IntegerNode:
	case *ast.BoolNode:
	case *ast.FloatNode:
	case *ast.UnaryNode:
		if t.Operator != "-" {
			return false
		}
		switch t.Node.(type) {
		case *ast.IntegerNode:
			return true
		case *ast.FloatNode:
			return true
		}

		return false
	case *ast.ArrayNode:
		for _, v := range t.Nodes {
			if !ExprIsLiteral(v) {
				return false
			}
		}
	default:
		return false
	}

	return true
}

func MarshalExprExpression(s string) graphql.Marshaler        { return graphql.MarshalString(s) }
func MarshalExprBooleanExpression(s string) graphql.Marshaler { return graphql.MarshalString(s) }
func MarshalExprStringExpression(s string) graphql.Marshaler  { return graphql.MarshalString(s) }

func exprExpressionWith(v interface{}, opts ...expr.Option) (string, error) {
	str, err := graphql.UnmarshalString(v)
	if err != nil {
		return "", err
	}

	_, err = expr.Compile(str, opts...)
	if err != nil {
		return "", validation.WrapError(err)
	}

	return str, nil
}

func UnmarshalExprExpression(v interface{}) (string, error) { return exprExpressionWith(v) }
func UnmarshalExprBooleanExpression(v interface{}) (string, error) {
	return exprExpressionWith(v, expr.AsBool())
}

func UnmarshalExprStringExpression(v interface{}) (string, error) {
	return exprExpressionWith(v, expr.AsKind(reflect.String))
}

func MarshalExprValue(n ast.Node) graphql.Marshaler      { return graphql.MarshalString(n.String()) }
func MarshalExprIdentifier(n ast.Node) graphql.Marshaler { return graphql.MarshalString(n.String()) }
func MarshalExprOperator(op string) graphql.Marshaler    { return graphql.MarshalString(op) }

func exprVal(v interface{}) (ast.Node, error) {
	str, err := graphql.UnmarshalString(v)
	if err != nil {
		return nil, err
	}

	t, err := parser.Parse(str)
	if err != nil {
		return nil, validation.WrapError(err)
	}

	return t.Node, nil
}

func UnmarshalExprValue(v interface{}) (ast.Node, error) {
	n, err := exprVal(v)
	if err != nil {
		return nil, validation.WrapError(err)
	}
	if !ExprIsLiteral(n) {
		return nil, validation.NewGenericError("must be a literal value")
	}

	return n, nil
}

func UnmarshalExprIdentifier(v interface{}) (ast.Node, error) {
	n, err := exprVal(v)
	if err != nil {
		return nil, validation.WrapError(err)
	}
	if !ExprIsID(n) {
		return nil, validation.NewGenericError("must be an identifier")
	}

	return n, nil
}

func UnmarshalExprOperator(v interface{}) (string, error) {
	n, err := exprVal(v)
	if err != nil {
		return "", validation.WrapError(err)
	}
	bin, ok := n.(*ast.BinaryNode)
	if !ok {
		return "", validation.NewGenericError("invalid operator")
	}
	if _, ok := bin.Left.(*ast.IdentifierNode); !ok {
		return "", validation.NewGenericError("invalid operator")
	}
	if _, ok := bin.Right.(*ast.IdentifierNode); !ok {
		return "", validation.NewGenericError("invalid operator")
	}

	return bin.Operator, nil
}

func UnmarshalExprStringMap(v interface{}) (map[string]string, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, validation.NewGenericError("must be a map")
	}
	res := make(map[string]string, len(m))
	for k, v := range m {
		str, err := UnmarshalExprStringExpression(v)
		if err != nil {
			return nil, MapValueError{Key: k, Err: err}
		}
		res[k] = str
	}

	return res, nil
}

func MarshalExprStringMap(v map[string]string) graphql.Marshaler {
	return graphql.MarshalAny(v)
}
