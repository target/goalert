package assignment

// Target contains information about the target, or assignee of an assignment.
type Target interface {
	TargetType() TargetType
	TargetID() string
}
type RawTarget struct {
	Type TargetType `json:"target_type"`
	ID   string     `json:"target_id"`
	Name string     `json:"target_name"`
}

func NewRawTarget(t Target) RawTarget {
	return RawTarget{Type: t.TargetType(), ID: t.TargetID()}
}
func (rt RawTarget) TargetType() TargetType {
	return rt.Type
}
func (rt RawTarget) TargetID() string {
	return rt.ID
}

// TargetName returns the name of the target. If unavailable, an empty string is returned.
func (rt RawTarget) TargetName() string {
	return rt.Name
}

// TargetNamer allows getting the friendly name of a target.
// Note: TargetName may return an empty string if the name is unavailable.
type TargetNamer interface {
	TargetName() string
}

type (
	// EscalationPolicyTarget implements the Target interface by wrapping an EscalationPolicy ID.
	EscalationPolicyTarget string
	// NotificationPolicyTarget implements the Target interface by wrapping a NotificationPolicy ID.
	NotificationPolicyTarget string
	// RotationTarget implements the Target interface by wrapping a Rotation ID.
	RotationTarget string
	// ServiceTarget implements the Target interface by wrapping a Service ID.
	ServiceTarget string
	// ScheduleTarget implements the Target interface by wrapping a Schedule ID.
	ScheduleTarget string
	// UserTarget implements the Target interface by wrapping a User ID.
	UserTarget string
	// NotificationChannelTarget implements the Target interface by wrapping a notification channel ID.
	NotificationChannelTarget string
	// IntegrationKeyTarget implements the Target interface by wrapping an IntegrationKey ID.
	IntegrationKeyTarget string
	// UserOverrideTarget implements the Target interface by wrapping an UserOverride ID.
	UserOverrideTarget string
	// ContactMethodTarget implements the Target interface by wrapping a ContactMethod ID.
	ContactMethodTarget string
	// NotificationRuleTarget implements the Target interface by wrapping an NotificationRule ID.
	NotificationRuleTarget string
)

// TargetType implements the Target interface.
func (EscalationPolicyTarget) TargetType() TargetType { return TargetTypeEscalationPolicy }

// TargetID implements the Target interface.
func (e EscalationPolicyTarget) TargetID() string { return string(e) }

// TargetType implements the Target interface.
func (NotificationPolicyTarget) TargetType() TargetType { return TargetTypeNotificationPolicy }

// TargetID implements the Target interface.
func (n NotificationPolicyTarget) TargetID() string { return string(n) }

// TargetType implements the Target interface.
func (RotationTarget) TargetType() TargetType { return TargetTypeRotation }

// TargetID implements the Target interface.
func (r RotationTarget) TargetID() string { return string(r) }

// TargetType implements the Target interface.
func (ServiceTarget) TargetType() TargetType { return TargetTypeService }

// TargetID implements the Target interface.
func (s ServiceTarget) TargetID() string { return string(s) }

// TargetType implements the Target interface.
func (ScheduleTarget) TargetType() TargetType { return TargetTypeSchedule }

// TargetID implements the Target interface.
func (s ScheduleTarget) TargetID() string { return string(s) }

// TargetType implements the Target interface.
func (UserTarget) TargetType() TargetType { return TargetTypeUser }

// TargetID implements the Target interface.
func (u UserTarget) TargetID() string { return string(u) }

// TargetType implements the Target interface.
func (NotificationChannelTarget) TargetType() TargetType { return TargetTypeNotificationChannel }

// TargetID implements the Target interface.
func (nc NotificationChannelTarget) TargetID() string { return string(nc) }

// TargetType implements the Target interface.
func (IntegrationKeyTarget) TargetType() TargetType { return TargetTypeIntegrationKey }

// TargetID implements the Target interface.
func (k IntegrationKeyTarget) TargetID() string { return string(k) }

// TargetType implements the Target interface.
func (UserOverrideTarget) TargetType() TargetType { return TargetTypeUserOverride }

// TargetID implements the Target interface.
func (k UserOverrideTarget) TargetID() string { return string(k) }

// TargetType implements the Target interface.
func (ContactMethodTarget) TargetType() TargetType { return TargetTypeContactMethod }

// TargetID implements the Target interface.
func (k ContactMethodTarget) TargetID() string { return string(k) }

// TargetType implements the Target interface.
func (NotificationRuleTarget) TargetType() TargetType { return TargetTypeNotificationRule }

// TargetID implements the Target interface.
func (k NotificationRuleTarget) TargetID() string { return string(k) }
