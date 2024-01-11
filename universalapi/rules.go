package universalapi

import (
	"context"
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/target/goalert/integrationkey/integrationkeyrule"
	"github.com/target/goalert/util/log"
)

type MatchOptions struct {
	SingleMatch bool //
}

// FetchRules returns a list of rules associated with the given intKey
func (h *Handler) FetchRules(ctx context.Context,
	intKey string) (rules []integrationkeyrule.Rule, err error) {

	rules, err = h.c.IntKeyRuleStore.FindManyByIntKey(ctx, h.c.DB, intKey)
	if err != nil {
		return nil, err
	}

	return
}

// MatchRules iterates through a list of rules running each rule template
// against the request body via expr returning rules which match.
// First match termination is achievable via the opts parameter.
func MatchRules(ctx context.Context, rules []integrationkeyrule.Rule, requestBody map[string]interface{}, opts *MatchOptions) []integrationkeyrule.Rule {
	matchedRules := []integrationkeyrule.Rule{}

	if opts == nil {
		opts = &MatchOptions{}
	}

	for _, rule := range rules {
		matched, err := matchExpressionWithExpr(rule.Filter, requestBody)
		if err != nil {
			log.Log(ctx, fmt.Errorf("evaluate rule: %s", err))
			continue
		}

		if matched {
			matchedRules = append(matchedRules, rule)
		}

		if len(matchedRules) != 0 && opts.SingleMatch {
			break
		}
	}

	return matchedRules
}

// matchExpressionWithExpr runs an expression evaluating it's boolean value with expr.
func matchExpressionWithExpr(expression string, env map[string]interface{}) (bool, error) {
	program, err := expr.Compile(expression)
	if err != nil {
		return false, err
	}

	result, err := expr.Run(program, env)
	if err != nil {
		return false, err
	}

	return result.(bool), nil
}
