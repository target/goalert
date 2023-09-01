package notifyapi

import (
	"fmt"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/pkg/errors"
	"github.com/ryanuber/go-glob"
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

type Rule struct {
	Filters []Filter
	Actions []Action
}

func FindMatchingRules(serviceID string, alertBody map[string]interface{}) (rules []Rule, err error) {
	allRules := findRulesByServiceID(serviceID)
	if len(allRules) == 0 {
		err = errors.Errorf("No rules for Service ID: '%s'", serviceID)
		return
	}

	for _, rule := range allRules {
		expression := ""
		for _, filter := range rule.Filters {
			field, operand, value := filter.Field, filter.Operand, filter.Value
			if containsEqualityOperator(operand) {
				var err error
				value, err = preprocessExpression(filter, alertBody)
				if err != nil {
					fmt.Printf("Error preprocessing rule: %v\n", err)
					continue
				}
			}
			if len(expression) == 0 {
				expression += fmt.Sprintf("%s %s %s", field, operand, value)
			} else {
				expression += fmt.Sprintf(" && %s %s %s", field, operand, value)
			}
		}

		matched, err := matchExpressionWithExpr(expression, alertBody)
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

func findRulesByServiceID(serviceID string) (rules []Rule) {
	rule := Rule{
		Filters: []Filter{
			{Field: "Temp", Operand: ">", Value: "25"},
			{Field: "message", Operand: "==", Value: "alert*DANGER"},
		},
		Actions: []Action{
			{Destination: Slack, Message: "ALERT!"},
		},
	}
	rules = []Rule{rule}
	return
}

func containsEqualityOperator(operator string) bool {
	return strings.Contains(operator, "=!") || strings.Contains(operator, "==")
}

func preprocessExpression(filter Filter, env map[string]interface{}) (string, error) {
	value, ok := env[filter.Field]
	if !ok {
		return "", fmt.Errorf("Field '%s' not found in environment", filter.Field)
	}

	stringValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("Field '%s' is not a string", filter.Field)
	}

	if glob.Glob(filter.Value, stringValue) {
		return fmt.Sprintf("'%s'", stringValue), nil
	}

	return fmt.Sprintf("'%s'", filter.Value), nil
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
