package graphql2

import (
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/schedule"
)

type OnCallNotificationRuleInput struct {
	schedule.OnCallNotificationRule
	Target assignment.RawTarget
}
