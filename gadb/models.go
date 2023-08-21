// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1

package gadb

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type EngineProcessingType string

const (
	EngineProcessingTypeEscalation   EngineProcessingType = "escalation"
	EngineProcessingTypeHeartbeat    EngineProcessingType = "heartbeat"
	EngineProcessingTypeNpCycle      EngineProcessingType = "np_cycle"
	EngineProcessingTypeRotation     EngineProcessingType = "rotation"
	EngineProcessingTypeSchedule     EngineProcessingType = "schedule"
	EngineProcessingTypeStatusUpdate EngineProcessingType = "status_update"
	EngineProcessingTypeVerify       EngineProcessingType = "verify"
	EngineProcessingTypeMessage      EngineProcessingType = "message"
	EngineProcessingTypeCleanup      EngineProcessingType = "cleanup"
	EngineProcessingTypeMetrics      EngineProcessingType = "metrics"
	EngineProcessingTypeCompat       EngineProcessingType = "compat"
)

func (e *EngineProcessingType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = EngineProcessingType(s)
	case string:
		*e = EngineProcessingType(s)
	default:
		return fmt.Errorf("unsupported scan type for EngineProcessingType: %T", src)
	}
	return nil
}

type NullEngineProcessingType struct {
	EngineProcessingType EngineProcessingType
	Valid                bool // Valid is true if EngineProcessingType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullEngineProcessingType) Scan(value interface{}) error {
	if value == nil {
		ns.EngineProcessingType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.EngineProcessingType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullEngineProcessingType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.EngineProcessingType), nil
}

type EnumAlertLogEvent string

const (
	EnumAlertLogEventCreated             EnumAlertLogEvent = "created"
	EnumAlertLogEventReopened            EnumAlertLogEvent = "reopened"
	EnumAlertLogEventStatusChanged       EnumAlertLogEvent = "status_changed"
	EnumAlertLogEventAssignmentChanged   EnumAlertLogEvent = "assignment_changed"
	EnumAlertLogEventEscalated           EnumAlertLogEvent = "escalated"
	EnumAlertLogEventClosed              EnumAlertLogEvent = "closed"
	EnumAlertLogEventNotificationSent    EnumAlertLogEvent = "notification_sent"
	EnumAlertLogEventResponseReceived    EnumAlertLogEvent = "response_received"
	EnumAlertLogEventAcknowledged        EnumAlertLogEvent = "acknowledged"
	EnumAlertLogEventPolicyUpdated       EnumAlertLogEvent = "policy_updated"
	EnumAlertLogEventDuplicateSuppressed EnumAlertLogEvent = "duplicate_suppressed"
	EnumAlertLogEventEscalationRequest   EnumAlertLogEvent = "escalation_request"
	EnumAlertLogEventNoNotificationSent  EnumAlertLogEvent = "no_notification_sent"
)

func (e *EnumAlertLogEvent) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = EnumAlertLogEvent(s)
	case string:
		*e = EnumAlertLogEvent(s)
	default:
		return fmt.Errorf("unsupported scan type for EnumAlertLogEvent: %T", src)
	}
	return nil
}

type NullEnumAlertLogEvent struct {
	EnumAlertLogEvent EnumAlertLogEvent
	Valid             bool // Valid is true if EnumAlertLogEvent is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullEnumAlertLogEvent) Scan(value interface{}) error {
	if value == nil {
		ns.EnumAlertLogEvent, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.EnumAlertLogEvent.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullEnumAlertLogEvent) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.EnumAlertLogEvent), nil
}

type EnumAlertLogSubjectType string

const (
	EnumAlertLogSubjectTypeUser             EnumAlertLogSubjectType = "user"
	EnumAlertLogSubjectTypeIntegrationKey   EnumAlertLogSubjectType = "integration_key"
	EnumAlertLogSubjectTypeHeartbeatMonitor EnumAlertLogSubjectType = "heartbeat_monitor"
	EnumAlertLogSubjectTypeChannel          EnumAlertLogSubjectType = "channel"
)

func (e *EnumAlertLogSubjectType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = EnumAlertLogSubjectType(s)
	case string:
		*e = EnumAlertLogSubjectType(s)
	default:
		return fmt.Errorf("unsupported scan type for EnumAlertLogSubjectType: %T", src)
	}
	return nil
}

type NullEnumAlertLogSubjectType struct {
	EnumAlertLogSubjectType EnumAlertLogSubjectType
	Valid                   bool // Valid is true if EnumAlertLogSubjectType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullEnumAlertLogSubjectType) Scan(value interface{}) error {
	if value == nil {
		ns.EnumAlertLogSubjectType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.EnumAlertLogSubjectType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullEnumAlertLogSubjectType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.EnumAlertLogSubjectType), nil
}

type EnumAlertSource string

const (
	EnumAlertSourceGrafana                EnumAlertSource = "grafana"
	EnumAlertSourceManual                 EnumAlertSource = "manual"
	EnumAlertSourceGeneric                EnumAlertSource = "generic"
	EnumAlertSourceEmail                  EnumAlertSource = "email"
	EnumAlertSourceSite24x7               EnumAlertSource = "site24x7"
	EnumAlertSourcePrometheusAlertmanager EnumAlertSource = "prometheusAlertmanager"
)

func (e *EnumAlertSource) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = EnumAlertSource(s)
	case string:
		*e = EnumAlertSource(s)
	default:
		return fmt.Errorf("unsupported scan type for EnumAlertSource: %T", src)
	}
	return nil
}

type NullEnumAlertSource struct {
	EnumAlertSource EnumAlertSource
	Valid           bool // Valid is true if EnumAlertSource is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullEnumAlertSource) Scan(value interface{}) error {
	if value == nil {
		ns.EnumAlertSource, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.EnumAlertSource.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullEnumAlertSource) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.EnumAlertSource), nil
}

type EnumAlertStatus string

const (
	EnumAlertStatusTriggered EnumAlertStatus = "triggered"
	EnumAlertStatusActive    EnumAlertStatus = "active"
	EnumAlertStatusClosed    EnumAlertStatus = "closed"
)

func (e *EnumAlertStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = EnumAlertStatus(s)
	case string:
		*e = EnumAlertStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for EnumAlertStatus: %T", src)
	}
	return nil
}

type NullEnumAlertStatus struct {
	EnumAlertStatus EnumAlertStatus
	Valid           bool // Valid is true if EnumAlertStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullEnumAlertStatus) Scan(value interface{}) error {
	if value == nil {
		ns.EnumAlertStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.EnumAlertStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullEnumAlertStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.EnumAlertStatus), nil
}

type EnumHeartbeatState string

const (
	EnumHeartbeatStateInactive  EnumHeartbeatState = "inactive"
	EnumHeartbeatStateHealthy   EnumHeartbeatState = "healthy"
	EnumHeartbeatStateUnhealthy EnumHeartbeatState = "unhealthy"
)

func (e *EnumHeartbeatState) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = EnumHeartbeatState(s)
	case string:
		*e = EnumHeartbeatState(s)
	default:
		return fmt.Errorf("unsupported scan type for EnumHeartbeatState: %T", src)
	}
	return nil
}

type NullEnumHeartbeatState struct {
	EnumHeartbeatState EnumHeartbeatState
	Valid              bool // Valid is true if EnumHeartbeatState is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullEnumHeartbeatState) Scan(value interface{}) error {
	if value == nil {
		ns.EnumHeartbeatState, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.EnumHeartbeatState.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullEnumHeartbeatState) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.EnumHeartbeatState), nil
}

type EnumIntegrationKeysType string

const (
	EnumIntegrationKeysTypeGrafana                EnumIntegrationKeysType = "grafana"
	EnumIntegrationKeysTypeGeneric                EnumIntegrationKeysType = "generic"
	EnumIntegrationKeysTypeEmail                  EnumIntegrationKeysType = "email"
	EnumIntegrationKeysTypeSite24x7               EnumIntegrationKeysType = "site24x7"
	EnumIntegrationKeysTypePrometheusAlertmanager EnumIntegrationKeysType = "prometheusAlertmanager"
)

func (e *EnumIntegrationKeysType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = EnumIntegrationKeysType(s)
	case string:
		*e = EnumIntegrationKeysType(s)
	default:
		return fmt.Errorf("unsupported scan type for EnumIntegrationKeysType: %T", src)
	}
	return nil
}

type NullEnumIntegrationKeysType struct {
	EnumIntegrationKeysType EnumIntegrationKeysType
	Valid                   bool // Valid is true if EnumIntegrationKeysType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullEnumIntegrationKeysType) Scan(value interface{}) error {
	if value == nil {
		ns.EnumIntegrationKeysType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.EnumIntegrationKeysType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullEnumIntegrationKeysType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.EnumIntegrationKeysType), nil
}

type EnumLimitType string

const (
	EnumLimitTypeNotificationRulesPerUser     EnumLimitType = "notification_rules_per_user"
	EnumLimitTypeContactMethodsPerUser        EnumLimitType = "contact_methods_per_user"
	EnumLimitTypeEpStepsPerPolicy             EnumLimitType = "ep_steps_per_policy"
	EnumLimitTypeEpActionsPerStep             EnumLimitType = "ep_actions_per_step"
	EnumLimitTypeParticipantsPerRotation      EnumLimitType = "participants_per_rotation"
	EnumLimitTypeRulesPerSchedule             EnumLimitType = "rules_per_schedule"
	EnumLimitTypeIntegrationKeysPerService    EnumLimitType = "integration_keys_per_service"
	EnumLimitTypeUnackedAlertsPerService      EnumLimitType = "unacked_alerts_per_service"
	EnumLimitTypeTargetsPerSchedule           EnumLimitType = "targets_per_schedule"
	EnumLimitTypeHeartbeatMonitorsPerService  EnumLimitType = "heartbeat_monitors_per_service"
	EnumLimitTypeUserOverridesPerSchedule     EnumLimitType = "user_overrides_per_schedule"
	EnumLimitTypeCalendarSubscriptionsPerUser EnumLimitType = "calendar_subscriptions_per_user"
)

func (e *EnumLimitType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = EnumLimitType(s)
	case string:
		*e = EnumLimitType(s)
	default:
		return fmt.Errorf("unsupported scan type for EnumLimitType: %T", src)
	}
	return nil
}

type NullEnumLimitType struct {
	EnumLimitType EnumLimitType
	Valid         bool // Valid is true if EnumLimitType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullEnumLimitType) Scan(value interface{}) error {
	if value == nil {
		ns.EnumLimitType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.EnumLimitType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullEnumLimitType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.EnumLimitType), nil
}

type EnumNotifChannelType string

const (
	EnumNotifChannelTypeSLACK          EnumNotifChannelType = "SLACK"
	EnumNotifChannelTypeWEBHOOK        EnumNotifChannelType = "WEBHOOK"
	EnumNotifChannelTypeSLACKUSERGROUP EnumNotifChannelType = "SLACK_USER_GROUP"
)

func (e *EnumNotifChannelType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = EnumNotifChannelType(s)
	case string:
		*e = EnumNotifChannelType(s)
	default:
		return fmt.Errorf("unsupported scan type for EnumNotifChannelType: %T", src)
	}
	return nil
}

type NullEnumNotifChannelType struct {
	EnumNotifChannelType EnumNotifChannelType
	Valid                bool // Valid is true if EnumNotifChannelType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullEnumNotifChannelType) Scan(value interface{}) error {
	if value == nil {
		ns.EnumNotifChannelType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.EnumNotifChannelType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullEnumNotifChannelType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.EnumNotifChannelType), nil
}

type EnumOutgoingMessagesStatus string

const (
	EnumOutgoingMessagesStatusPending        EnumOutgoingMessagesStatus = "pending"
	EnumOutgoingMessagesStatusSending        EnumOutgoingMessagesStatus = "sending"
	EnumOutgoingMessagesStatusQueuedRemotely EnumOutgoingMessagesStatus = "queued_remotely"
	EnumOutgoingMessagesStatusSent           EnumOutgoingMessagesStatus = "sent"
	EnumOutgoingMessagesStatusDelivered      EnumOutgoingMessagesStatus = "delivered"
	EnumOutgoingMessagesStatusFailed         EnumOutgoingMessagesStatus = "failed"
	EnumOutgoingMessagesStatusBundled        EnumOutgoingMessagesStatus = "bundled"
)

func (e *EnumOutgoingMessagesStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = EnumOutgoingMessagesStatus(s)
	case string:
		*e = EnumOutgoingMessagesStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for EnumOutgoingMessagesStatus: %T", src)
	}
	return nil
}

type NullEnumOutgoingMessagesStatus struct {
	EnumOutgoingMessagesStatus EnumOutgoingMessagesStatus
	Valid                      bool // Valid is true if EnumOutgoingMessagesStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullEnumOutgoingMessagesStatus) Scan(value interface{}) error {
	if value == nil {
		ns.EnumOutgoingMessagesStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.EnumOutgoingMessagesStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullEnumOutgoingMessagesStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.EnumOutgoingMessagesStatus), nil
}

type EnumOutgoingMessagesType string

const (
	EnumOutgoingMessagesTypeAlertNotification          EnumOutgoingMessagesType = "alert_notification"
	EnumOutgoingMessagesTypeVerificationMessage        EnumOutgoingMessagesType = "verification_message"
	EnumOutgoingMessagesTypeTestNotification           EnumOutgoingMessagesType = "test_notification"
	EnumOutgoingMessagesTypeAlertStatusUpdate          EnumOutgoingMessagesType = "alert_status_update"
	EnumOutgoingMessagesTypeAlertNotificationBundle    EnumOutgoingMessagesType = "alert_notification_bundle"
	EnumOutgoingMessagesTypeAlertStatusUpdateBundle    EnumOutgoingMessagesType = "alert_status_update_bundle"
	EnumOutgoingMessagesTypeScheduleOnCallNotification EnumOutgoingMessagesType = "schedule_on_call_notification"
)

func (e *EnumOutgoingMessagesType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = EnumOutgoingMessagesType(s)
	case string:
		*e = EnumOutgoingMessagesType(s)
	default:
		return fmt.Errorf("unsupported scan type for EnumOutgoingMessagesType: %T", src)
	}
	return nil
}

type NullEnumOutgoingMessagesType struct {
	EnumOutgoingMessagesType EnumOutgoingMessagesType
	Valid                    bool // Valid is true if EnumOutgoingMessagesType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullEnumOutgoingMessagesType) Scan(value interface{}) error {
	if value == nil {
		ns.EnumOutgoingMessagesType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.EnumOutgoingMessagesType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullEnumOutgoingMessagesType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.EnumOutgoingMessagesType), nil
}

type EnumRotationType string

const (
	EnumRotationTypeMonthly EnumRotationType = "monthly"
	EnumRotationTypeWeekly  EnumRotationType = "weekly"
	EnumRotationTypeDaily   EnumRotationType = "daily"
	EnumRotationTypeHourly  EnumRotationType = "hourly"
)

func (e *EnumRotationType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = EnumRotationType(s)
	case string:
		*e = EnumRotationType(s)
	default:
		return fmt.Errorf("unsupported scan type for EnumRotationType: %T", src)
	}
	return nil
}

type NullEnumRotationType struct {
	EnumRotationType EnumRotationType
	Valid            bool // Valid is true if EnumRotationType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullEnumRotationType) Scan(value interface{}) error {
	if value == nil {
		ns.EnumRotationType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.EnumRotationType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullEnumRotationType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.EnumRotationType), nil
}

type EnumSwitchoverState string

const (
	EnumSwitchoverStateIdle       EnumSwitchoverState = "idle"
	EnumSwitchoverStateInProgress EnumSwitchoverState = "in_progress"
	EnumSwitchoverStateUseNextDb  EnumSwitchoverState = "use_next_db"
)

func (e *EnumSwitchoverState) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = EnumSwitchoverState(s)
	case string:
		*e = EnumSwitchoverState(s)
	default:
		return fmt.Errorf("unsupported scan type for EnumSwitchoverState: %T", src)
	}
	return nil
}

type NullEnumSwitchoverState struct {
	EnumSwitchoverState EnumSwitchoverState
	Valid               bool // Valid is true if EnumSwitchoverState is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullEnumSwitchoverState) Scan(value interface{}) error {
	if value == nil {
		ns.EnumSwitchoverState, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.EnumSwitchoverState.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullEnumSwitchoverState) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.EnumSwitchoverState), nil
}

type EnumThrottleType string

const (
	EnumThrottleTypeNotifications  EnumThrottleType = "notifications"
	EnumThrottleTypeNotifications2 EnumThrottleType = "notifications_2"
)

func (e *EnumThrottleType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = EnumThrottleType(s)
	case string:
		*e = EnumThrottleType(s)
	default:
		return fmt.Errorf("unsupported scan type for EnumThrottleType: %T", src)
	}
	return nil
}

type NullEnumThrottleType struct {
	EnumThrottleType EnumThrottleType
	Valid            bool // Valid is true if EnumThrottleType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullEnumThrottleType) Scan(value interface{}) error {
	if value == nil {
		ns.EnumThrottleType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.EnumThrottleType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullEnumThrottleType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.EnumThrottleType), nil
}

type EnumUserContactMethodType string

const (
	EnumUserContactMethodTypePUSH    EnumUserContactMethodType = "PUSH"
	EnumUserContactMethodTypeEMAIL   EnumUserContactMethodType = "EMAIL"
	EnumUserContactMethodTypeVOICE   EnumUserContactMethodType = "VOICE"
	EnumUserContactMethodTypeSMS     EnumUserContactMethodType = "SMS"
	EnumUserContactMethodTypeWEBHOOK EnumUserContactMethodType = "WEBHOOK"
	EnumUserContactMethodTypeSLACKDM EnumUserContactMethodType = "SLACK_DM"
)

func (e *EnumUserContactMethodType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = EnumUserContactMethodType(s)
	case string:
		*e = EnumUserContactMethodType(s)
	default:
		return fmt.Errorf("unsupported scan type for EnumUserContactMethodType: %T", src)
	}
	return nil
}

type NullEnumUserContactMethodType struct {
	EnumUserContactMethodType EnumUserContactMethodType
	Valid                     bool // Valid is true if EnumUserContactMethodType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullEnumUserContactMethodType) Scan(value interface{}) error {
	if value == nil {
		ns.EnumUserContactMethodType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.EnumUserContactMethodType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullEnumUserContactMethodType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.EnumUserContactMethodType), nil
}

type EnumUserRole string

const (
	EnumUserRoleUnknown EnumUserRole = "unknown"
	EnumUserRoleUser    EnumUserRole = "user"
	EnumUserRoleAdmin   EnumUserRole = "admin"
)

func (e *EnumUserRole) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = EnumUserRole(s)
	case string:
		*e = EnumUserRole(s)
	default:
		return fmt.Errorf("unsupported scan type for EnumUserRole: %T", src)
	}
	return nil
}

type NullEnumUserRole struct {
	EnumUserRole EnumUserRole
	Valid        bool // Valid is true if EnumUserRole is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullEnumUserRole) Scan(value interface{}) error {
	if value == nil {
		ns.EnumUserRole, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.EnumUserRole.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullEnumUserRole) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.EnumUserRole), nil
}

type Alert struct {
	ID              int64
	ServiceID       uuid.NullUUID
	Source          EnumAlertSource
	Status          EnumAlertStatus
	EscalationLevel int32
	LastEscalation  sql.NullTime
	LastProcessed   sql.NullTime
	CreatedAt       time.Time
	DedupKey        sql.NullString
	Summary         string
	Details         string
}

type AlertFeedback struct {
	AlertID     int64
	NoiseReason string
}

type AlertLog struct {
	ID                  int64
	AlertID             sql.NullInt64
	Timestamp           sql.NullTime
	Event               EnumAlertLogEvent
	Message             string
	SubType             NullEnumAlertLogSubjectType
	SubUserID           uuid.NullUUID
	SubIntegrationKeyID uuid.NullUUID
	SubClassifier       string
	Meta                pqtype.NullRawMessage
	SubHbMonitorID      uuid.NullUUID
	SubChannelID        uuid.NullUUID
}

type AlertMetric struct {
	ID          int64
	AlertID     int64
	ServiceID   uuid.UUID
	TimeToAck   sql.NullInt64
	TimeToClose sql.NullInt64
	Escalated   bool
	ClosedAt    time.Time
}

type AlertStatusSubscription struct {
	ID              int64
	ChannelID       uuid.NullUUID
	ContactMethodID uuid.NullUUID
	AlertID         int64
	LastAlertStatus EnumAlertStatus
}

type AuthBasicUser struct {
	UserID       uuid.UUID
	Username     string
	PasswordHash string
	ID           int64
}

type AuthLinkRequest struct {
	ID         uuid.UUID
	ProviderID string
	SubjectID  string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	Metadata   json.RawMessage
}

type AuthNonce struct {
	ID        uuid.UUID
	CreatedAt time.Time
}

type AuthSubject struct {
	ProviderID string
	SubjectID  string
	UserID     uuid.UUID
	ID         int64
	CmID       uuid.NullUUID
}

type AuthUserSession struct {
	ID           uuid.UUID
	CreatedAt    time.Time
	UserAgent    string
	UserID       uuid.NullUUID
	LastAccessAt time.Time
}

type Config struct {
	ID        int32
	Schema    int32
	Data      []byte
	CreatedAt time.Time
}

type ConfigLimit struct {
	ID  EnumLimitType
	Max int32
}

type EngineProcessingVersion struct {
	TypeID  EngineProcessingType
	Version int32
	State   json.RawMessage
}

type EpStepOnCallUser struct {
	UserID    uuid.UUID
	EpStepID  uuid.UUID
	StartTime time.Time
	EndTime   sql.NullTime
	ID        int64
}

type EscalationPolicy struct {
	ID          uuid.UUID
	Name        string
	Description string
	Repeat      int32
	StepCount   int32
}

type EscalationPolicyAction struct {
	ID                     uuid.UUID
	EscalationPolicyStepID uuid.UUID
	UserID                 uuid.NullUUID
	ScheduleID             uuid.NullUUID
	RotationID             uuid.NullUUID
	ChannelID              uuid.NullUUID
}

type EscalationPolicyState struct {
	EscalationPolicyID         uuid.UUID
	EscalationPolicyStepID     uuid.NullUUID
	EscalationPolicyStepNumber int32
	AlertID                    int64
	LastEscalation             sql.NullTime
	LoopCount                  int32
	ForceEscalation            bool
	ServiceID                  uuid.UUID
	NextEscalation             sql.NullTime
	ID                         int64
}

type EscalationPolicyStep struct {
	ID                 uuid.UUID
	Delay              int32
	StepNumber         int32
	EscalationPolicyID uuid.UUID
}

type GorpMigration struct {
	ID        string
	AppliedAt sql.NullTime
}

type HeartbeatMonitor struct {
	ID                uuid.UUID
	Name              string
	ServiceID         uuid.UUID
	HeartbeatInterval int64
	LastState         EnumHeartbeatState
	LastHeartbeat     sql.NullTime
}

type IntegrationKey struct {
	ID        uuid.UUID
	Name      string
	Type      EnumIntegrationKeysType
	ServiceID uuid.UUID
}

type Keyring struct {
	ID               string
	VerificationKeys []byte
	SigningKey       []byte
	NextKey          []byte
	NextRotation     sql.NullTime
	RotationCount    int64
}

type Label struct {
	ID           int64
	TgtServiceID uuid.UUID
	Key          string
	Value        string
}

type NotificationChannel struct {
	ID        uuid.UUID
	CreatedAt time.Time
	Type      EnumNotifChannelType
	Name      string
	Value     string
	Meta      json.RawMessage
}

type NotificationPolicyCycle struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	AlertID     int32
	RepeatCount int32
	StartedAt   time.Time
	Checked     bool
	LastTick    sql.NullTime
}

type OutgoingMessage struct {
	ID                     uuid.UUID
	MessageType            EnumOutgoingMessagesType
	ContactMethodID        uuid.NullUUID
	CreatedAt              time.Time
	LastStatus             EnumOutgoingMessagesStatus
	LastStatusAt           sql.NullTime
	StatusDetails          string
	FiredAt                sql.NullTime
	SentAt                 sql.NullTime
	RetryCount             int32
	NextRetryAt            sql.NullTime
	SendingDeadline        sql.NullTime
	UserID                 uuid.NullUUID
	AlertID                sql.NullInt64
	CycleID                uuid.NullUUID
	ServiceID              uuid.NullUUID
	EscalationPolicyID     uuid.NullUUID
	AlertLogID             sql.NullInt64
	UserVerificationCodeID uuid.NullUUID
	ProviderMsgID          sql.NullString
	ProviderSeq            int32
	ChannelID              uuid.NullUUID
	StatusAlertIds         []int64
	ScheduleID             uuid.NullUUID
	SrcValue               sql.NullString
}

type RegionID struct {
	Name string
	ID   int32
}

type Rotation struct {
	ID               uuid.UUID
	Name             string
	Description      string
	Type             EnumRotationType
	StartTime        time.Time
	ShiftLength      int64
	TimeZone         string
	LastProcessed    sql.NullTime
	ParticipantCount int32
}

type RotationParticipant struct {
	ID         uuid.UUID
	RotationID uuid.UUID
	Position   int32
	UserID     uuid.UUID
}

type RotationState struct {
	RotationID            uuid.UUID
	Position              int32
	RotationParticipantID uuid.UUID
	ShiftStart            time.Time
	ID                    int64
	Version               int32
}

type Schedule struct {
	ID            uuid.UUID
	Name          string
	Description   string
	TimeZone      string
	LastProcessed sql.NullTime
}

type ScheduleDatum struct {
	ScheduleID    uuid.UUID
	LastCleanupAt sql.NullTime
	Data          json.RawMessage
	ID            int64
}

type ScheduleOnCallUser struct {
	ScheduleID uuid.UUID
	StartTime  time.Time
	EndTime    sql.NullTime
	UserID     uuid.UUID
	ID         int64
}

type ScheduleRule struct {
	ID            uuid.UUID
	ScheduleID    uuid.UUID
	Sunday        bool
	Monday        bool
	Tuesday       bool
	Wednesday     bool
	Thursday      bool
	Friday        bool
	Saturday      bool
	StartTime     time.Time
	EndTime       time.Time
	CreatedAt     time.Time
	TgtUserID     uuid.NullUUID
	TgtRotationID uuid.NullUUID
	IsActive      bool
}

type Service struct {
	ID                   uuid.UUID
	Name                 string
	Description          string
	EscalationPolicyID   uuid.UUID
	MaintenanceExpiresAt sql.NullTime
}

type SwitchoverLog struct {
	ID        int64
	Timestamp time.Time
	Data      json.RawMessage
}

type SwitchoverState struct {
	Ok           bool
	CurrentState EnumSwitchoverState
	DbID         uuid.UUID
}

type TwilioSmsCallback struct {
	PhoneNumber string
	CallbackID  uuid.UUID
	Code        int32
	ID          int64
	SentAt      time.Time
	AlertID     sql.NullInt64
	ServiceID   uuid.NullUUID
}

type TwilioSmsError struct {
	PhoneNumber  string
	ErrorMessage string
	Outgoing     bool
	OccurredAt   time.Time
	ID           int64
}

type TwilioVoiceError struct {
	PhoneNumber  string
	ErrorMessage string
	Outgoing     bool
	OccurredAt   time.Time
	ID           int64
}

type User struct {
	ID                            uuid.UUID
	Bio                           string
	Email                         string
	Role                          EnumUserRole
	Name                          string
	AvatarUrl                     string
	AlertStatusLogContactMethodID uuid.NullUUID
}

type UserCalendarSubscription struct {
	ID         uuid.UUID
	Name       string
	UserID     uuid.UUID
	LastAccess sql.NullTime
	LastUpdate time.Time
	CreatedAt  time.Time
	Disabled   bool
	ScheduleID uuid.UUID
	Config     json.RawMessage
}

type UserContactMethod struct {
	ID                  uuid.UUID
	Name                string
	Type                EnumUserContactMethodType
	Value               string
	Disabled            bool
	UserID              uuid.UUID
	LastTestVerifyAt    sql.NullTime
	Metadata            pqtype.NullRawMessage
	EnableStatusUpdates bool
	Pending             bool
}

type UserFavorite struct {
	UserID                uuid.UUID
	TgtServiceID          uuid.NullUUID
	ID                    int64
	TgtRotationID         uuid.NullUUID
	TgtScheduleID         uuid.NullUUID
	TgtEscalationPolicyID uuid.NullUUID
	TgtUserID             uuid.NullUUID
}

type UserNotificationRule struct {
	ID              uuid.UUID
	DelayMinutes    int32
	ContactMethodID uuid.UUID
	UserID          uuid.UUID
	CreatedAt       sql.NullTime
}

type UserOverride struct {
	ID            uuid.UUID
	StartTime     time.Time
	EndTime       time.Time
	AddUserID     uuid.NullUUID
	RemoveUserID  uuid.NullUUID
	TgtScheduleID uuid.UUID
}

type UserSlackDatum struct {
	ID          uuid.UUID
	AccessToken string
}

type UserVerificationCode struct {
	ID              uuid.UUID
	Code            int32
	ExpiresAt       time.Time
	ContactMethodID uuid.UUID
	Sent            bool
}
