package graphqlapp

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
	"github.com/pkg/errors"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/validation"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Expr App

func (a *App) Expr() graphql2.ExprResolver                    { return (*Expr)(a) }
func (q *Query) Expr(context.Context) (*graphql2.Expr, error) { return &graphql2.Expr{}, nil }

var errTooComplex = errors.New("expression is too complex")

func gqlErrTooComplex(ctx context.Context) error {
	return &gqlerror.Error{
		Message: errTooComplex.Error(),
		Path:    graphql.GetPath(ctx),
		Extensions: map[string]interface{}{
			"code": graphql2.ErrorCodeExprTooComplex,
		},
	}
}

var supportedOperators = []string{"==", "!=", "<", ">", "<=", ">=", "in"}

func isID(n ast.Node) bool {
	switch t := n.(type) {
	case *ast.IdentifierNode:
		return true
	case *ast.MemberNode:
		return !t.Method && isID(t.Node) && (isID(t.Property) || isLiteral(t.Property))
	case *ast.ChainNode:
		return isID(t.Node)
	}

	return false
}

func isLiteral(n ast.Node) bool {
	switch t := n.(type) {
	case *ast.StringNode:
	case *ast.IntegerNode:
	case *ast.BoolNode:
	case *ast.FloatNode:
	case *ast.UnaryNode:
		return isLiteral(t.Node)
	case *ast.ArrayNode:
		for _, v := range t.Nodes {
			if !isLiteral(v) {
				return false
			}
		}
	default:
		return false
	}

	return true
}

func containsClause(n *ast.BinaryNode) (clause *graphql2.Clause, ok bool) {
	call, ok := n.Left.(*ast.BuiltinNode)
	if !ok {
		return nil, false
	}
	un, ok := n.Right.(*ast.UnaryNode)
	if !ok || un.Operator != "-" {
		return nil, false
	}
	val, ok := un.Node.(*ast.IntegerNode)
	if !ok || val.Value != 1 {
		return nil, false
	}

	if call.Name != "indexOf" {
		return nil, false
	}
	if len(call.Arguments) != 2 {
		return nil, false
	}

	if !isID(call.Arguments[0]) {
		return nil, false
	}

	if !isLiteral(call.Arguments[1]) {
		return nil, false
	}

	var op string
	switch n.Operator {
	case "==":
		op = "not_contains"
	case "!=":
		op = "contains"
	default:
		return nil, false
	}

	return &graphql2.Clause{
		Field:    call.Arguments[0].String(),
		Operator: op,
		Value:    litToJSON(call.Arguments[1]),
	}, true
}

func litToJSON(n ast.Node) string {
	if !isLiteral(n) {
		panic("not a literal")
	}

	if s, ok := n.(*ast.StringNode); ok {
		data, err := json.Marshal(s.Value)
		if err != nil {
			panic(err) // shouldn't be possible
		}
		return string(data)
	}

	ary, ok := n.(*ast.ArrayNode)
	if !ok {
		return n.String()
	}
	var vals []string
	for _, v := range ary.Nodes {
		vals = append(vals, litToJSON(v))
	}

	return fmt.Sprintf("[%s]", strings.Join(vals, ","))
}

func exprToCondition(expr string) (*graphql2.Condition, error) {
	tree, err := parser.Parse(expr)
	if err != nil {
		return nil, err
	}

	top, ok := tree.Node.(*ast.BinaryNode)
	if !ok {
		return nil, errTooComplex
	}

	var clauses []graphql2.Clause
	var handleBinary func(n *ast.BinaryNode) error
	handleBinary = func(n *ast.BinaryNode) error {
		if clause, ok := containsClause(n); ok {
			fmt.Println("CONTAINS CLAUSE:", clause)
			clauses = append(clauses, *clause)
			return nil
		}
		if n.Operator == "and" || n.Operator == "&&" {
			l, ok := n.Left.(*ast.BinaryNode)
			if !ok {
				return errTooComplex
			}
			r, ok := n.Right.(*ast.BinaryNode)
			if !ok {
				return errTooComplex
			}
			if err := handleBinary(l); err != nil {
				return err
			}
			if err := handleBinary(r); err != nil {
				return err
			}
			return nil
		}

		if !slices.Contains(supportedOperators, n.Operator) {
			return errTooComplex
		}

		if !isID(n.Left) {
			return errTooComplex
		}

		if !isLiteral(n.Right) {
			return errTooComplex
		}

		clauses = append(clauses, graphql2.Clause{
			Field:    n.Left.String(),
			Operator: n.Operator,
			Value:    litToJSON(n.Right),
		})

		return nil
	}

	if err := handleBinary(top); err != nil {
		return nil, err
	}

	return &graphql2.Condition{
		Clauses: clauses,
	}, nil
}

func (e *Expr) ExprToCondition(ctx context.Context, _ *graphql2.Expr, input graphql2.ExprToConditionInput) (*graphql2.Condition, error) {
	cond, err := exprToCondition(input.Expr)
	if errors.Is(err, errTooComplex) {
		return nil, gqlErrTooComplex(ctx)
	}
	if err != nil {
		addInputError(ctx, validation.NewFieldError("input.expr", err.Error()))
		return nil, errAlreadySet
	}

	return cond, nil
}

func jsonToExprValue(val string) (ast.Node, error) {
	var v any
	err := json.Unmarshal([]byte(val), &v)
	if err != nil {
		return nil, err
	}

	n, err := goToExpr(v)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func goToExpr(val any) (ast.Node, error) {
	switch t := val.(type) {
	case string:
		return &ast.StringNode{Value: t}, nil
	case float64:
		return &ast.FloatNode{Value: t}, nil
	case bool:
		return &ast.BoolNode{Value: t}, nil
	case []interface{}:
		var nodes []ast.Node
		for _, v := range t {
			n, err := goToExpr(v)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, n)
		}

		return &ast.ArrayNode{Nodes: nodes}, nil
	}

	return nil, errors.New("invalid value")
}

func (e *Expr) ConditionToExpr(ctx context.Context, _ *graphql2.Expr, input graphql2.ConditionToExprInput) (string, error) {
	var exprs []string
	for i, c := range input.Condition.Clauses {
		path := fmt.Sprintf("input.condition[%d]", i)

		left, err := parser.Parse(c.Field)
		if err != nil {
			addInputError(ctx, validation.NewFieldError(path+".field", "invalid field"))
			return "", errAlreadySet
		}
		if !isID(left.Node) {
			addInputError(ctx, validation.NewFieldError(path+".field", "invalid field"))
			return "", errAlreadySet
		}

		rightNode, err := jsonToExprValue(c.Value)
		if err != nil {
			addInputError(ctx, validation.NewFieldError(path+".value", err.Error()))
			return "", errAlreadySet
		}

		if c.Operator == "contains" || c.Operator == "not_contains" {
			op := "!="
			if c.Operator == "not_contains" {
				op = "=="
			}
			exprs = append(exprs, fmt.Sprintf("indexOf(%s, %s) %s -1", left.Node.String(), rightNode.String(), op))
			continue
		}

		if !slices.Contains(supportedOperators, c.Operator) {
			addInputError(ctx, validation.NewFieldError(path+".operator", "unsupported operator"))
			return "", errAlreadySet
		}

		if _, ok := left.Node.(*ast.IdentifierNode); !ok {
			addInputError(ctx, validation.NewFieldError(path+".field", "invalid field"))
			return "", errAlreadySet
		}

		exprs = append(exprs, fmt.Sprintf("%s %s %s", left.Node.String(), c.Operator, rightNode.String()))
	}

	return strings.Join(exprs, " and "), nil
}
