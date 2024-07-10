package smoke

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/expflag"
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
			Input struct {
				Dest struct {
					Type string `json:"type"`
					Args params `json:"args"`
				} `json:"dest"`
				Params params `json:"params"`
			} `json:"input"`
		}
		vars.Input.Params = dyn
		vars.Input.Dest.Type = destType
		vars.Input.Dest.Args = dest

		return *h.GraphQLQueryUserVarsT(t, harness.DefaultGraphQLAdminUserID, actionValidQuery, "TestActionValid", vars)
	}

	res := check("invalid", params{}, params{})
	if assert.Len(t, res.Errors, 1) {
		assert.EqualValues(t, "actionInputValidate.input.dest.type", res.Errors[0].Path)
		assert.Equal(t, "unsupported destination type: invalid", res.Errors[0].Message)
		assert.Equal(t, "INVALID_INPUT_VALUE", res.Errors[0].Extensions.Code)
	}

	res = check("builtin-alert", params{}, params{"invalid-expr": `foo+`})
	if assert.Len(t, res.Errors, 1) {
		assert.EqualValues(t, "actionInputValidate.input.params", res.Errors[0].Path)
		assert.Contains(t, res.Errors[0].Message, "unexpected token")
		assert.Equal(t, "INVALID_MAP_FIELD_VALUE", res.Errors[0].Extensions.Code)
		assert.Equal(t, "invalid-expr", res.Errors[0].Extensions.Key)
	}
}
