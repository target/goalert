package uik

import (
	"fmt"

	"github.com/expr-lang/expr/vm"
	"github.com/target/goalert/gadb"
)

// CompiledConfig is a compiled version of a UIKConfigV1.
type CompiledConfig struct {
	gadb.UIKConfigV1
	CompiledRules  []CompiledRule
	DefaultActions []CompiledAction
}

// RuleError is an error that occurred while processing a rule.
type RuleError struct {
	Index int
	Name  string
	Err   error
}

func (r *RuleError) Error() string {
	return fmt.Sprintf("rule %d (%s): %s", r.Index, r.Name, r.Err)
}

// NewCompiledConfig will compile a UIKConfigV1 into a CompiledConfig.
func NewCompiledConfig(cfg gadb.UIKConfigV1) (*CompiledConfig, error) {
	res := &CompiledConfig{
		UIKConfigV1:    cfg,
		CompiledRules:  make([]CompiledRule, len(cfg.Rules)),
		DefaultActions: make([]CompiledAction, len(cfg.DefaultActions)),
	}
	for i, r := range cfg.Rules {
		p, err := NewCompiledRule(r)
		if err != nil {
			return nil, &RuleError{
				Index: i,
				Name:  r.Name,
				Err:   fmt.Errorf("compile rules: %w", err),
			}
		}
		res.CompiledRules[i] = *p
	}
	for i, a := range cfg.DefaultActions {
		c, err := NewCompiledAction(a)
		if err != nil {
			return nil, &ActionError{
				Index: i,
				Err:   fmt.Errorf("compile default actions: %w", err),
			}
		}
		res.DefaultActions[i] = *c
	}
	return res, nil
}

// Run will execute the compiled config against the provided VM and environment.
func (c *CompiledConfig) Run(vm *vm.VM, env any) (actions []gadb.UIKActionV1, err error) {
	var anyMatched bool
	for i, p := range c.CompiledRules {
		ruleActions, matched, err := p.Run(vm, env)
		if err != nil {
			return nil, &RuleError{
				Index: i,
				Err:   fmt.Errorf("run rules: %w", err),
			}
		}
		actions = append(actions, ruleActions...)
		anyMatched = anyMatched || matched
		if matched && !p.ContinueAfterMatch {
			break
		}
	}

	if anyMatched {
		return actions, nil
	}

	act, err := runActions(vm, c.DefaultActions, env)
	if err != nil {
		return nil, fmt.Errorf("run default actions: %w", err)
	}

	return act, nil
}
