package uik

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/target/goalert/integrationkey"
)

// BuildRuleExpr will build an expression string for a rule.
//
// The resulting expression will return nil unless the condition is true, in
// which case it will return the actions as an array of objects, where each
// object is a map of dynamic parameters, key being the parameter ID and value
// being the result of the dynamic parameter expression.
//
// All compared values are converted to strings before comparison to ensure
// proper type handling.
//
// The expected format is: `string(<condition>) == "true" ? [{ <action-1-param-1>: string(<action-1-param-1-expr>), ... }] : nil`
// See ExampleBuildRuleExpr for an example.
func BuildRuleExpr(cond string, actions []integrationkey.Action) string {
	var actionExpr []string
	for _, a := range actions {
		var params []string
		for k, v := range a.DynamicParams {
			params = append(params, fmt.Sprintf(`%s: string(%s)`, strconv.Quote(k), v))
		}

		actionExpr = append(actionExpr, fmt.Sprintf(`{ %s }`, strings.Join(params, ", ")))
	}

	src := "[" + strings.Join(actionExpr, ",") + "]"
	if cond != "" {
		src = fmt.Sprintf(`string(%s) == "true" ? %s : nil`, cond, src)
	}

	return src
}

// CompileRule returns an Expr-compiled rule.
func CompileRule(cond string, actions []integrationkey.Action) (*vm.Program, error) {
	return expr.Compile(BuildRuleExpr(cond, actions), expr.AsKind(reflect.Slice), expr.AllowUndefinedVariables(), expr.Optimize(true))
}
