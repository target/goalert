package graphql2

import (
	"encoding/json"
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

func MarshalJSONValue(v any) graphql.Marshaler {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return graphql.MarshalString(string(data))
}

func UnmarshalJSONValue(v any) (any, error) {
	str, err := graphql.UnmarshalString(v)
	if err != nil {
		return nil, err
	}
	var data any
	err = json.Unmarshal([]byte(str), &data)
	if err != nil {
		return nil, validation.WrapError(err)
	}
	return data, nil
}

func MarshalExprExpression(s string) graphql.Marshaler {
	return graphql.MarshalString(s)
}

func UnmarshalExprExpression(v interface{}) (string, error) {
	str, err := graphql.UnmarshalString(v)
	if err != nil {
		return "", err
	}

	_, err = expr.Compile(str)
	if err != nil {
		return "", validation.WrapError(err)
	}

	return str, nil
}

func MarshalExprBooleanExpression(s string) graphql.Marshaler {
	return graphql.MarshalString(s)
}

func UnmarshalExprBooleanExpression(v interface{}) (string, error) {
	str, err := graphql.UnmarshalString(v)
	if err != nil {
		return "", err
	}

	_, err = expr.Compile(str, expr.AsBool())
	if err != nil {
		return "", validation.WrapError(err)
	}

	return str, nil
}

func MarshalExprStringExpression(s string) graphql.Marshaler {
	return graphql.MarshalString(s)
}

func UnmarshalExprStringExpression(v interface{}) (string, error) {
	str, err := graphql.UnmarshalString(v)
	if err != nil {
		return "", err
	}

	_, err = expr.Compile(str, expr.AsKind(reflect.String))
	if err != nil {
		return "", validation.WrapError(err)
	}

	return str, nil
}

func MarshalExprValue(n ast.Node) graphql.Marshaler {
	return graphql.MarshalString(n.String())
}

func UnmarshalExprValue(v interface{}) (ast.Node, error) {
	str, err := graphql.UnmarshalString(v)
	if err != nil {
		return nil, err
	}

	t, err := parser.Parse(str)
	if err != nil {
		return nil, validation.WrapError(err)
	}
	if !ExprIsLiteral(t.Node) {
		return nil, validation.NewGenericError("must be a literal value")
	}

	return t.Node, nil
}

func MarshalExprIdentifier(n ast.Node) graphql.Marshaler {
	return graphql.MarshalString(n.String())
}

func UnmarshalExprIdentifier(v interface{}) (ast.Node, error) {
	str, err := graphql.UnmarshalString(v)
	if err != nil {
		return nil, err
	}

	t, err := parser.Parse(str)
	if err != nil {
		return nil, validation.WrapError(err)
	}
	if !ExprIsID(t.Node) {
		return nil, validation.NewGenericError("must be an identifier")
	}

	return t.Node, nil
}

func MarshalExprOperator(op string) graphql.Marshaler {
	return graphql.MarshalString(op)
}

func UnmarshalExprOperator(v interface{}) (string, error) {
	str, err := graphql.UnmarshalString(v)
	if err != nil {
		return "", err
	}

	t, err := parser.Parse("a " + str + " b")
	if err != nil {
		return "", validation.NewGenericError("invalid operator")
	}
	bin, ok := t.Node.(*ast.BinaryNode)
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
