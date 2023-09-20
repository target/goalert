package notifyapi

import (
	"context"
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/service/rule"
)

type Destination string

const (
	Slack Destination = "slack"
	Email Destination = "email"
	SMS   Destination = "sms"
)

type Filter struct {
	Field   string
	Operand string
	Value   string
}

type Action struct {
	Destination Destination
	Message     string
}

func (h *Handler) FindMatchingRules(ctx context.Context, intKeyID string, alertBody map[string]interface{}) (rules []rule.Rule, err error) {
	allRules, err := h.c.ServiceRuleStore.GetRulesForIntegrationKey(ctx, permission.ServiceID(ctx), intKeyID)
	if err != nil {
		return nil, err
	}
	if len(allRules) == 0 {
		return
	}

	for _, rule := range allRules {
		matched, err := matchExpressionWithExpr(rule.FilterString, alertBody)
		if err != nil {
			fmt.Printf("Error evaluating rule: %v\n", err)
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
