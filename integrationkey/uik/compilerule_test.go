package uik

import (
	"fmt"

	"github.com/target/goalert/gadb"
)

func ExampleBuildRuleExpr() {
	actions := []gadb.UIKActionV1{
		{
			Params: map[string]string{
				"param-1": "req.body.param1",
			},
		},
	}
	fmt.Println(BuildRuleExpr("req.body.shouldSend", actions))
	// Output: string(req.body.shouldSend) == "true" ? [{ "param-1": string(req.body.param1) }] : nil
}
