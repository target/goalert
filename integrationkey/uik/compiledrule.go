package uik

import (
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/target/goalert/gadb"
)

// CompiledRule is a compiled version of a UIKRuleV1.
type CompiledRule struct {
	gadb.UIKRuleV1
	Condition *vm.Program
	Actions   []CompiledAction
}

// ActionError is an error that occurred while processing an action.
type ActionError struct {
	Index int
	Err   error
}

func (a *ActionError) Error() string {
	return fmt.Sprintf("action %d: %s", a.Index, a.Err)
}

// ConditionError is an error that occurred while processing a condition.
type ConditionError struct {
	Err error
}

func (c *ConditionError) Error() string {
	return fmt.Sprintf("condition: %s", c.Err)
}

// NewCompiledRule will compile a UIKRuleV1 into a CompiledRule.
func NewCompiledRule(r gadb.UIKRuleV1) (*CompiledRule, error) {
	cond, err := expr.Compile(r.ConditionExpr, expr.AllowUndefinedVariables(), expr.Optimize(true), expr.AsBool())
	if err != nil {
		return nil, &ConditionError{
			Err: fmt.Errorf("compile: %w", err),
		}
	}

	act := make([]CompiledAction, len(r.Actions))
	for i, a := range r.Actions {
		c, err := NewCompiledAction(a)
		if err != nil {
			return nil, &ActionError{
				Index: i,
				Err:   fmt.Errorf("compile: %w", err),
			}
		}
		act[i] = *c
	}

	return &CompiledRule{
		UIKRuleV1: r,
		Condition: cond,
		Actions:   act,
	}, nil
}

// Run will execute the compiled rule against the provided VM and environment.
func (r *CompiledRule) Run(vm *vm.VM, env any) (actions []gadb.UIKActionV1, matched bool, err error) {
	res, err := vm.Run(r.Condition, env)
	if err != nil {
		return nil, false, &ConditionError{
			Err: fmt.Errorf("run: %w", err),
		}
	}
	if !res.(bool) {
		return nil, false, nil
	}

	actions, err = runActions(vm, r.Actions, env)
	return actions, true, err
}

func runActions(vm *vm.VM, actions []CompiledAction, env any) (result []gadb.UIKActionV1, err error) {
	result = make([]gadb.UIKActionV1, len(actions))
	for i, a := range actions {
		res, err := a.Run(vm, env)
		if err != nil {
			return nil, &ActionError{
				Index: i,
				Err:   fmt.Errorf("run: %w", err),
			}
		}
		result[i] = res
	}
	return result, nil
}
