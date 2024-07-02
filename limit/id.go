package limit

import "github.com/target/goalert/validation/validate"

// ID represents the identifier for a given system limit.
type ID string

// IDs of configurable limits.
const (
	// Maximum number of notification rules per user.
	NotificationRulesPerUser ID = "notification_rules_per_user"
	// Maximum number of contact methods per user.
	ContactMethodsPerUser ID = "contact_methods_per_user"
	// Maximum number of steps on a single escalation policy.
	EPStepsPerPolicy ID = "ep_steps_per_policy"
	// Maximum number of actions on a single escalation policy step.
	EPActionsPerStep ID = "ep_actions_per_step"
	// Maximum number of participants per rotation.
	ParticipantsPerRotation ID = "participants_per_rotation"
	// Pertains to all rules for all assignments/targets.
	RulesPerSchedule ID = "rules_per_schedule"
	// Maximum number of integration keys per service.
	IntegrationKeysPerService ID = "integration_keys_per_service"
	// Only affects newly created alerts but not re-escalated ones.
	UnackedAlertsPerService ID = "unacked_alerts_per_service"
	// Maximum number of targets per schedule.
	TargetsPerSchedule ID = "targets_per_schedule"
	// Maximum number of heartbeat monitors per service.
	HeartbeatMonitorsPerService ID = "heartbeat_monitors_per_service"
	// Only limits future overrides (i.e. end in the future).
	UserOverridesPerSchedule ID = "user_overrides_per_schedule"
	// Maximum number of calendar subscriptions per user.
	CalendarSubscriptionsPerUser ID = "calendar_subscriptions_per_user"
	// Maximum number of pending signals per service.
	PendingSignalsPerService ID = "pending_signals_per_service"
	// Maximum number of pending signals per destination per service.
	PendingSignalsPerDestPerService ID = "pending_signals_per_dest_per_service"
)

// Valid returns nil if a given ID is valid, a validation error is returned otherwise.
func (id ID) Valid() error {
	return validate.OneOf("LimitID", id,
		NotificationRulesPerUser,
		ContactMethodsPerUser,
		EPStepsPerPolicy,
		EPActionsPerStep,
		ParticipantsPerRotation,
		RulesPerSchedule,
		IntegrationKeysPerService,
		UnackedAlertsPerService,
		TargetsPerSchedule,
		HeartbeatMonitorsPerService,
		UserOverridesPerSchedule,
		CalendarSubscriptionsPerUser,
		PendingSignalsPerService,
		PendingSignalsPerDestPerService,
	)
}
