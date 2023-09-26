package notifyapi

import (
	"context"
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/service/rule"
	"github.com/target/goalert/util/log"
)

// findMatchingRules returns all rules associated with the given integration key whose
// filters match the given requestBody
func (h *Handler) findMatchingRules(ctx context.Context, intKeyID string, requestBody map[string]interface{}) (rules []rule.Rule, err error) {
	allRules, err := h.c.ServiceRuleStore.GetRulesForIntegrationKey(ctx, permission.ServiceID(ctx), intKeyID)
	if err != nil {
		return nil, err
	}

	for _, rule := range allRules {
		matched, err := matchExpressionWithExpr(rule.FilterString, requestBody)
		if err != nil {
			ctx := log.WithField(ctx, "RuleID", rule.ID)
			log.Log(ctx, fmt.Errorf("evaluate rule: %s", err))
			continue
		}

		if matched {
			rules = append(rules, rule)
		}
	}
	return
}

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
