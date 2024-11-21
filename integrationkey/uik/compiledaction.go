package uik

import (
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/target/goalert/gadb"
)

// CompiledAction is a compiled version of a UIKActionV1.
type CompiledAction struct {
	gadb.UIKActionV1
	Params map[string]*vm.Program
}

// ParamError is an error that occurred while processing a param.
type ParamError struct {
	ParamID string
	Err     error
}

func (p *ParamError) Error() string {
	return fmt.Sprintf("param %s: %s", p.ParamID, p.Err)
}

// NewCompiledAction will compile a UIKActionV1 into a CompiledAction.
func NewCompiledAction(a gadb.UIKActionV1) (*CompiledAction, error) {
	res := &CompiledAction{
		UIKActionV1: a,
		Params:      make(map[string]*vm.Program, len(a.Params)),
	}
	for k, v := range a.Params {
		p, err := expr.Compile(v, expr.AllowUndefinedVariables(), expr.Optimize(true))
		if err != nil {
			return nil, &ParamError{
				ParamID: k,
				Err:     fmt.Errorf("compile: %w", err),
			}
		}
		res.Params[k] = p
	}
	return res, nil
}

// Run will execute the compiled action against the provided VM and environment.
func (a *CompiledAction) Run(vm *vm.VM, env any) (result gadb.UIKActionV1, err error) {
	result.ChannelID = a.ChannelID
	result.Dest = a.Dest
	result.Params = make(map[string]string, len(a.Params))
	for k, v := range a.Params {
		res, err := vm.Run(v, env)
		if err != nil {
			return result, &ParamError{
				ParamID: k,
				Err:     fmt.Errorf("run: %w", err),
			}
		}
		result.Params[k] = fmt.Sprintf("%v", res)
	}
	return result, nil
}
