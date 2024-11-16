package uik

import (
	"testing"

	"github.com/expr-lang/expr/vm"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/gadb"
)

func TestCompiledRule_Run(t *testing.T) {
	r := gadb.UIKRuleV1{
		ConditionExpr: "shouldRun",
		Actions: []gadb.UIKActionV1{
			{Params: map[string]string{"key": "someVar"}},
		},
	}

	var vm vm.VM
	compiledRule, err := NewCompiledRule(r)
	require.NoError(t, err, "should compile a valid rule")
	actions, matched, err := compiledRule.Run(&vm, map[string]any{"shouldRun": false})
	require.NoError(t, err, "should run a valid rule")
	require.Empty(t, actions, "should not return actions if condition evals false")
	require.False(t, matched, "should not return matched if condition evals false")

	actions, matched, err = compiledRule.Run(&vm, map[string]any{"shouldRun": true, "someVar": "value"})
	require.NoError(t, err, "should run a valid rule")
	require.True(t, matched, "should return matched if condition evals true")
	require.Len(t, actions, 1, "should run a valid rule")
	require.Equal(t, "value", actions[0].Params["key"], "should evaluate action params")
}
