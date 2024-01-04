package universalapi

import (
	"context"
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/target/goalert/util/log"
)

// Temporary implementation, should be later expanded to support multiple "actions"
type Rule struct {
	Action    string // close or create alert
	Dedup     string
	Template  string
	SendAlert bool
}

type SearchOptions struct {
	FindMultiple bool // continue fetching rules beyond first found
}

// FetchRules returns a list of rules associated with the given intKey
// stopping at the first found by default. The opts param can be used to
// fetch all matching rules by setting FindMultiple to true.
func (h *Handler) FetchRules(ctx context.Context, intKey string, opts *SearchOptions) (rules []Rule, err error) {
	// if opts == nil {
	// 	opts = &SearchOptions{}
	// }

	// serviceID := permission.ServiceID(ctx)

	// if opts.FindMultiple {
	// 	rules, err = h.c.ServiceRuleStore.FindManyByIntegrationKey(ctx, serviceID, intKey)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// } else {
	// 	rules, err = h.c.ServiceRuleStore.FindOneByIntegrationKey(ctx, serviceID, intKey)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	return
}

// MatchRules iterates through a list of rules running each rule template
// against the request body via expr returning rules which match.
func MatchRules(ctx context.Context, rules []Rule, requestBody map[string]interface{}) []Rule {
	matchedRules := []Rule{}

	for _, rule := range rules {
		matched, err := matchExpressionWithExpr(rule.Template, requestBody)
		if err != nil {
			log.Log(ctx, fmt.Errorf("evaluate rule: %s", err))
			continue
		}

		if matched {
			matchedRules = append(matchedRules, rule)
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
