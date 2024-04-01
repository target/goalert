package graphqlapp

import (
	"testing"

	"github.com/expr-lang/expr/parser"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/graphql2"
)

func TestIsID(t *testing.T) {
	check := func(desc, expr string, expected bool) {
		t.Helper()

		t.Run(desc, func(t *testing.T) {
			t.Log("expr:", expr)
			tree, err := parser.Parse(expr)
			require.NoError(t, err)

			require.Equal(t, expected, isID(tree.Node))
		})
	}

	check("simple", "foo", true)
	check("member", "foo.bar", true)
	check("method", "foo.bar()", false)
	check("array member", "foo[0].bar.baz", true)
	check("map member", "foo['asdf'].bar.baz", true)
	check("map member", "foo['asdf'].bar.baz", true)
	check("maybe member", "foo['asdf']?.bar.baz", true)
}

func TestExprToCondition(t *testing.T) {
	check := func(desc, expr string, expected []graphql2.Clause) {
		t.Helper()

		t.Run(desc, func(t *testing.T) {
			t.Log("expr:", expr)
			cond, err := exprToCondition(expr)
			require.NoError(t, err)
			require.NotNil(t, cond)
			require.Equal(t, expected, cond.Clauses)
		})
	}

	check("simple", "expr == true", []graphql2.Clause{
		{Field: "expr", Operator: "==", Value: "true"},
	})

	check("member", "expr[0].bar == true", []graphql2.Clause{
		{Field: "expr[0].bar", Operator: "==", Value: "true"},
	})

	check("string", "expr == \"true\"", []graphql2.Clause{
		{Field: "expr", Operator: "==", Value: "\"true\""},
	})
	check("multi", "expr == true && expr2 == 1 and expr contains \"yep\" and expr not contains 'asdf'", []graphql2.Clause{
		{Field: "expr", Operator: "==", Value: "true"},
		{Field: "expr2", Operator: "==", Value: "1"},
		{Field: "expr", Operator: "contains", Value: "\"yep\""},
		{Field: "expr", Operator: "contains", Value: "\"asdf\"", Negate: true},
	})

	check("one of", "expr in ['a', 'b', 'c'] and expr not in ['d']", []graphql2.Clause{
		{Field: "expr", Operator: "in", Value: `["a","b","c"]`},
		{Field: "expr", Operator: "in", Value: `["d"]`, Negate: true},
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

	check("simple", graphql2.ClauseInput{Field: "foo", Operator: "==", Value: "true"}, "foo == true")
	check("string", graphql2.ClauseInput{Field: "foo", Operator: "==", Value: "\"true\""}, "foo == \"true\"")
	check("array", graphql2.ClauseInput{Field: "foo", Operator: "in", Value: `["a","b","c"]`}, `foo in ["a", "b", "c"]`)
	check("contains", graphql2.ClauseInput{Field: "foo", Operator: "contains", Value: "\"asdf\""}, `foo contains "asdf"`)
	check("not contains", graphql2.ClauseInput{Field: "foo", Operator: "contains", Value: "\"asdf\"", Negate: true}, `foo not contains "asdf"`)

	t.Run("invalid", func(t *testing.T) {
		_, err := clauseToExpr("", graphql2.ClauseInput{Field: "foo", Operator: "contains", Value: "asdf"})
		require.Error(t, err)
	})
}
