package graphql2

import (
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/schedule"
)

type OnCallNotificationRuleInput struct {
	schedule.OnCallNotificationRule
	Dest gadb.DestV1
}
