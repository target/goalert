// Code generated by devtools/limitapigen DO NOT EDIT.

package graphql2

import (
	"fmt"

	"github.com/target/goalert/limit"
	"github.com/target/goalert/validation"
)

// MapLimitValues will map a Limit struct into a flat list of SystemLimit structs.
func MapLimitValues(l limit.Limits) []SystemLimit {
	return []SystemLimit{
		{ID: "CalendarSubscriptionsPerUser", Description: "Maximum number of calendar subscriptions per user.", Value: l[limit.CalendarSubscriptionsPerUser]},
		{ID: "ContactMethodsPerUser", Description: "Maximum number of contact methods per user.", Value: l[limit.ContactMethodsPerUser]},
		{ID: "EPActionsPerStep", Description: "Maximum number of actions on a single escalation policy step.", Value: l[limit.EPActionsPerStep]},
		{ID: "EPStepsPerPolicy", Description: "Maximum number of steps on a single escalation policy.", Value: l[limit.EPStepsPerPolicy]},
		{ID: "HeartbeatMonitorsPerService", Description: "Maximum number of heartbeat monitors per service.", Value: l[limit.HeartbeatMonitorsPerService]},
		{ID: "IntegrationKeysPerService", Description: "Maximum number of integration keys per service.", Value: l[limit.IntegrationKeysPerService]},
		{ID: "NotificationRulesPerUser", Description: "Maximum number of notification rules per user.", Value: l[limit.NotificationRulesPerUser]},
		{ID: "ParticipantsPerRotation", Description: "Maximum number of participants per rotation.", Value: l[limit.ParticipantsPerRotation]},
		{ID: "PendingSignalsPerDestPerService", Description: "Maximum number of pending signals per destination per service.", Value: l[limit.PendingSignalsPerDestPerService]},
		{ID: "PendingSignalsPerService", Description: "Maximum number of pending signals per service.", Value: l[limit.PendingSignalsPerService]},
		{ID: "RulesPerSchedule", Description: "Pertains to all rules for all assignments/targets.", Value: l[limit.RulesPerSchedule]},
		{ID: "TargetsPerSchedule", Description: "Maximum number of targets per schedule.", Value: l[limit.TargetsPerSchedule]},
		{ID: "UnackedAlertsPerService", Description: "Only affects newly created alerts but not re-escalated ones.", Value: l[limit.UnackedAlertsPerService]},
		{ID: "UserOverridesPerSchedule", Description: "Only limits future overrides (i.e. end in the future).", Value: l[limit.UserOverridesPerSchedule]},
	}
}

// ApplyLimitValues will apply a list of LimitValues to a Limit struct.
func ApplyLimitValues(l limit.Limits, vals []SystemLimitInput) (limit.Limits, error) {
	for _, v := range vals {
		switch v.ID {
		case "CalendarSubscriptionsPerUser":
			l[limit.CalendarSubscriptionsPerUser] = v.Value
		case "ContactMethodsPerUser":
			l[limit.ContactMethodsPerUser] = v.Value
		case "EPActionsPerStep":
			l[limit.EPActionsPerStep] = v.Value
		case "EPStepsPerPolicy":
			l[limit.EPStepsPerPolicy] = v.Value
		case "HeartbeatMonitorsPerService":
			l[limit.HeartbeatMonitorsPerService] = v.Value
		case "IntegrationKeysPerService":
			l[limit.IntegrationKeysPerService] = v.Value
		case "NotificationRulesPerUser":
			l[limit.NotificationRulesPerUser] = v.Value
		case "ParticipantsPerRotation":
			l[limit.ParticipantsPerRotation] = v.Value
		case "PendingSignalsPerDestPerService":
			l[limit.PendingSignalsPerDestPerService] = v.Value
		case "PendingSignalsPerService":
			l[limit.PendingSignalsPerService] = v.Value
		case "RulesPerSchedule":
			l[limit.RulesPerSchedule] = v.Value
		case "TargetsPerSchedule":
			l[limit.TargetsPerSchedule] = v.Value
		case "UnackedAlertsPerService":
			l[limit.UnackedAlertsPerService] = v.Value
		case "UserOverridesPerSchedule":
			l[limit.UserOverridesPerSchedule] = v.Value
		default:
			return l, validation.NewFieldError("ID", fmt.Sprintf("unknown limit ID '%s'", v.ID))
		}
	}
	return l, nil
}
