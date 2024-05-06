package graphqlapp

import (
	"context"
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

		if !graphql2.ExprIsID(n.Left) {
			return errTooComplex
		}

		if !graphql2.ExprIsLiteral(n.Right) {
			return errTooComplex
		}

		clauses = append(clauses, graphql2.Clause{
			Field:    n.Left,
			Operator: n.Operator,
			Value:    n.Right,
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
	cond, err := exprToCondition(string(input.Expr))
	if errors.Is(err, errTooComplex) {
		return nil, gqlErrTooComplex(ctx)
	}
	if err != nil {
		addInputError(ctx, validation.NewFieldError("input.expr", err.Error()))
		return nil, errAlreadySet
	}

	return cond, nil
}

func clauseToExpr(path string, c graphql2.ClauseInput) (string, error) {
	if !slices.Contains(supportedOperators, c.Operator) {
		return "", validation.NewFieldError(path+".operator", "unsupported operator")
	}

	var negateStr string
	if c.Negate {
		negateStr = "not "
	}
	return fmt.Sprintf("%s %s%s %s", c.Field.String(), negateStr, c.Operator, c.Value), nil
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
