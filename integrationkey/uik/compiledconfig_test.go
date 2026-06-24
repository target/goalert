package uik

import (
	"testing"

	"github.com/expr-lang/expr/vm"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/gadb"
)

func TestCompiledConfig_Run(t *testing.T) {
	cfg := gadb.UIKConfigV1{
		Rules: []gadb.UIKRuleV1{
			{
				Name:          "rule1",
				ConditionExpr: "shouldRun1",
				Actions: []gadb.UIKActionV1{
					// Lets use a string literal to simplify this test.
					{Params: map[string]string{"key": `"value1"`}},
				},
			},
			{
				Name:               "rule2",
				ConditionExpr:      "shouldRun2",
				ContinueAfterMatch: true,
				Actions: []gadb.UIKActionV1{
					{Params: map[string]string{"key": `"value2"`}},
				},
			},
			{
				Name:          "rule3",
				ConditionExpr: "shouldRun3",
				Actions: []gadb.UIKActionV1{
					{Params: map[string]string{"key": `"value3"`}},
				},
			},
		},
		DefaultActions: []gadb.UIKActionV1{
			{Params: map[string]string{"key": `"valueDefault"`}},
		},
	}
	cmp, err := NewCompiledConfig(cfg)
	require.NoError(t, err, "should compile a valid config")
	var vm vm.VM

	check := func(desc string, env any, expValues []string) {
		t.Helper()
		t.Run(desc, func(t *testing.T) {
			res, err := cmp.Run(&vm, env)
			require.NoError(t, err, "should run a valid config")
			require.Len(t, res, len(expValues), "should evaluate all actions")

			for i, a := range res {
				require.Equal(t, expValues[i], a.Params["key"], "should evaluate action params")
			}
		})
	}

	check("default actions",
		// since no rules match, we should only see the default action
		map[string]any{
			"shouldRun1": false,
			"shouldRun2": false,
			"shouldRun3": false,
		},
		[]string{"valueDefault"})

	check("stop at first match",
		// by default, processing stops at first match
		map[string]any{
			"shouldRun1": true,
			"shouldRun2": true,
			"shouldRun3": true,
		},
		[]string{"value1"})

	check("stop at first match",
		// rule #2 has ContinueAfterMatch set to true, so rule #3 should also run
		map[string]any{
			"shouldRun1": false,
			"shouldRun2": true,
			"shouldRun3": true,
		},
		[]string{"value2", "value3"})
}
