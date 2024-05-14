package graphql2

import (
	"testing"

	"github.com/expr-lang/expr/parser"
	"github.com/stretchr/testify/require"
)

func TestExprIsID(t *testing.T) {
	check := func(desc, expr string, expected bool) {
		t.Helper()

		t.Run(desc, func(t *testing.T) {
			t.Log("expr:", expr)
			tree, err := parser.Parse(expr)
			require.NoError(t, err)

			require.Equal(t, expected, ExprIsID(tree.Node))
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
