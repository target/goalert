package smoke

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/expflag"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/test/smoke/harness"
)

const actionValidQuery = `
query TestActionValid($input: ActionInput!) {
	actionInputValidate(input: $input)
}
`

// TestActionValid tests the action validation query.
func TestActionValid(t *testing.T) {
	t.Parallel()

	h := harness.NewHarnessWithFlags(t, "", "", expflag.FlagSet{expflag.UnivKeys})
	defer h.Close()

	type params map[string]string
	check := func(destType string, dest, dyn params) harness.QLResponse {
		t.Helper()
		var vars struct {
			Input graphql2.ActionInput `json:"input"`
		}
		vars.Input.Params = []graphql2.DynamicParamInput{}
		for k, v := range dyn {
			vars.Input.Params = append(vars.Input.Params, graphql2.DynamicParamInput{ParamID: k, Expr: v})
		}
		vars.Input.Dest = &graphql2.DestinationInput{Type: destType, Values: []graphql2.FieldValueInput{}}
		for k, v := range dest {
			vars.Input.Dest.Values = append(vars.Input.Dest.Values, graphql2.FieldValueInput{FieldID: k, Value: v})
		}

		return *h.GraphQLQueryUserVarsT(t, harness.DefaultGraphQLAdminUserID, actionValidQuery, "TestActionValid", vars)
	}

	res := check("invalid", params{}, params{})
	if assert.Len(t, res.Errors, 1) {
		assert.EqualValues(t, "actionInputValidate.input.dest.type", res.Errors[0].Path)
		assert.Equal(t, "unsupported destination type", res.Errors[0].Message)
		assert.Equal(t, "INVALID_INPUT_VALUE", res.Errors[0].Extensions.Code)
	}

	res = check("builtin-alert", params{}, params{"invalid-expr": `foo+`})
	if assert.Len(t, res.Errors, 1) {
		assert.EqualValues(t, "actionInputValidate.input.params.0.expr", res.Errors[0].Path)
		assert.Contains(t, res.Errors[0].Message, "unexpected token")
		assert.Equal(t, "", res.Errors[0].Extensions.Code)
	}
}
