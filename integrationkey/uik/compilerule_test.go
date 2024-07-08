package uik

import (
	"fmt"

	"github.com/target/goalert/integrationkey"
)

func ExampleBuildRuleExpr() {
	actions := []integrationkey.Action{
		{
			DynamicParams: map[string]string{
				"param-1": "req.body.param1",
			},
		},
	}
	fmt.Println(BuildRuleExpr("req.body.shouldSend", actions))
	// Output: string(req.body.shouldSend) == "true" ? [{ "param-1": string(req.body.param1) }] : nil
}
