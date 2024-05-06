package graphqlapp

import (
	"testing"

	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/graphql2"
)

func TestExprToCondition(t *testing.T) {
	check := func(desc, expr string, expected []graphql2.Clause) {
		t.Helper()

		t.Run(desc, func(t *testing.T) {
			t.Log("expr:", expr)
			cond, err := exprToCondition(expr)
			require.NoError(t, err)
			require.NotNil(t, cond)
			require.Len(t, cond.Clauses, len(expected))
			for i := range expected {
				require.Equal(t, expected[i].Field.String(), cond.Clauses[i].Field.String(), "clause[%d].Field", i)
				require.Equal(t, expected[i].Operator, cond.Clauses[i].Operator, "clause[%d].Operator", i)
				require.Equal(t, expected[i].Value.String(), cond.Clauses[i].Value.String(), "clause[%d].Value", i)
				require.Equal(t, expected[i].Negate, cond.Clauses[i].Negate, "clause[%d].Negate", i)
			}
		})
	}
	expr := func(s string) ast.Node {
		t.Helper()

		tree, err := parser.Parse(s)
		require.NoError(t, err)
		return tree.Node
	}

	check("simple", "expr == true", []graphql2.Clause{
		{Field: expr("expr"), Operator: "==", Value: expr("true")},
	})

	check("member", "expr[0].bar == true", []graphql2.Clause{
		{Field: expr("expr[0].bar"), Operator: "==", Value: expr("true")},
	})

	check("string", "expr == \"true\"", []graphql2.Clause{
		{Field: expr("expr"), Operator: "==", Value: expr("\"true\"")},
	})
	check("multi", "expr == true && expr2 == 1 and expr contains \"yep\" and expr not contains 'asdf'", []graphql2.Clause{
		{Field: expr("expr"), Operator: "==", Value: expr("true")},
		{Field: expr("expr2"), Operator: "==", Value: expr("1")},
		{Field: expr("expr"), Operator: "contains", Value: expr("\"yep\"")},
		{Field: expr("expr"), Operator: "contains", Value: expr("\"asdf\""), Negate: true},
	})

	check("one of", "expr in ['a', 'b', 'c'] and expr not in ['d']", []graphql2.Clause{
		{Field: expr("expr"), Operator: "in", Value: expr(`["a","b","c"]`)},
		{Field: expr("expr"), Operator: "in", Value: expr(`["d"]`), Negate: true},
	})

	t.Run("too complex", func(t *testing.T) {
		_, err := exprToCondition("1+1 == 2")
		require.ErrorIs(t, err, errTooComplex)
	})
	t.Run("invalid", func(t *testing.T) {
		_, err := exprToCondition("1 + asd\"f")
		require.Error(t, err)
	})
}

func TestClauseToExpr(t *testing.T) {
	check := func(desc string, clause graphql2.ClauseInput, expected string) {
		t.Helper()
		t.Run(desc, func(t *testing.T) {
			t.Log("clause:", clause)
			expr, err := clauseToExpr("", clause)
			require.NoError(t, err)
			require.Equal(t, expected, expr)
		})
	}
	expr := func(s string) ast.Node {
		t.Helper()

		tree, err := parser.Parse(s)
		require.NoError(t, err)
		return tree.Node
	}

	check("simple", graphql2.ClauseInput{Field: expr("foo"), Operator: "==", Value: expr("true")}, "foo == true")
	check("string", graphql2.ClauseInput{Field: expr("foo"), Operator: "==", Value: expr("\"true\"")}, "foo == \"true\"")
	check("array", graphql2.ClauseInput{Field: expr("foo"), Operator: "in", Value: expr(`["a","b","c"]`)}, `foo in ["a", "b", "c"]`)
	check("contains", graphql2.ClauseInput{Field: expr("foo"), Operator: "contains", Value: expr("\"asdf\"")}, `foo contains "asdf"`)
	check("not contains", graphql2.ClauseInput{Field: expr("foo"), Operator: "contains", Value: expr("\"asdf\""), Negate: true}, `foo not contains "asdf"`)
}
