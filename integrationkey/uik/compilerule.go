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

func CompileRule(cond string, actions []integrationkey.Action) (*vm.Program, error) {
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

	return expr.Compile(src, expr.AsKind(reflect.Slice), expr.AllowUndefinedVariables(), expr.Optimize(true))
}
