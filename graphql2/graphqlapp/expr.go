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

var supportedOperators = []string{"==", "!=", "<", ">", "<=", ">=", "in", "contains", "matches"}

// isID returns true if the node is a valid identifier.
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

// isLiteral returns true if the node is a literal value (scalar or array of scalar values).
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

// litToJSON converts a literal node to a JSON string.
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

// getBinaryNode returns the binary node and whether it should be negated.
func getBinaryNode(n ast.Node) (node *ast.BinaryNode, negate bool) {
	if un, ok := n.(*ast.UnaryNode); ok && un.Operator == "not" {
		negate = true
		node, _ = un.Node.(*ast.BinaryNode)
	} else {
		node, _ = n.(*ast.BinaryNode)
	}

	return node, negate
}

// exprToCondition converts an expression string to a Condition.
func exprToCondition(expr string) (*graphql2.Condition, error) {
	tree, err := parser.Parse(expr)
	if err != nil {
		return nil, err
	}

	top, topNegate := getBinaryNode(tree.Node)
	if top == nil {
		return nil, errTooComplex
	}

	var clauses []graphql2.Clause
	var handleBinary func(n *ast.BinaryNode, negate bool) error
	handleBinary = func(n *ast.BinaryNode, negate bool) error {
		if n.Operator == "and" || n.Operator == "&&" { // AND, process left hand side first, then right hand side
			if negate {
				// This would require inverting the remaining expression which
				// would equal to an OR operation which is not supported.
				return errTooComplex
			}

			left, leftNegate := getBinaryNode(n.Left)
			right, rightNegate := getBinaryNode(n.Right)
			if left == nil || right == nil {
				return errTooComplex
			}

			if err := handleBinary(left, leftNegate); err != nil {
				return err
			}
			if err := handleBinary(right, rightNegate); err != nil {
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
			Negate:   negate,
		})

		return nil
	}

	if err := handleBinary(top, topNegate); err != nil {
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

func clauseToExpr(path string, c graphql2.ClauseInput) (string, error) {

	left, err := parser.Parse(c.Field)
	if err != nil {
		return "", validation.NewFieldError(path+".field", "invalid field")
	}
	if !isID(left.Node) {
		return "", validation.NewFieldError(path+".field", "invalid field")
	}

	rightNode, err := jsonToExprValue(c.Value)
	if err != nil {
		return "", validation.NewFieldError(path+".value", err.Error())
	}

	if !slices.Contains(supportedOperators, c.Operator) {
		return "", validation.NewFieldError(path+".operator", "unsupported operator")
	}

	if _, ok := left.Node.(*ast.IdentifierNode); !ok {
		return "", validation.NewFieldError(path+".field", "invalid field")
	}

	var negateStr string
	if c.Negate {
		negateStr = "not "
	}
	return fmt.Sprintf("%s %s%s %s", left.Node.String(), negateStr, c.Operator, rightNode.String()), nil
}

func (e *Expr) ConditionToExpr(ctx context.Context, _ *graphql2.Expr, input graphql2.ConditionToExprInput) (string, error) {
	var exprs []string
	for i, c := range input.Condition.Clauses {
		str, err := clauseToExpr(fmt.Sprintf("input.condition[%d]", i), c)
		if err != nil {
			addInputError(ctx, err)
			return "", errAlreadySet
		}

		exprs = append(exprs, str)
	}

	return strings.Join(exprs, " and "), nil
}
