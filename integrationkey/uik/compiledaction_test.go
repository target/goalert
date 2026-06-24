package uik

import (
	"testing"

	"github.com/expr-lang/expr/vm"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/gadb"
)

func TestCompiledAction_Run(t *testing.T) {
	act := gadb.UIKActionV1{Params: map[string]string{"key": "someVar"}}

	var vm vm.VM
	compiledAct, err := NewCompiledAction(act)
	require.NoError(t, err, "should compile a valid action")
	res, err := compiledAct.Run(&vm, map[string]any{"someVar": "value"})
	require.NoError(t, err, "should run a valid action")
	require.Equal(t, "value", res.Params["key"], "should run a valid action")
}
