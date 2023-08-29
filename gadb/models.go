// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0

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
	EngineProcessingTypeCleanup      EngineProcessingType = "cleanup"
	EngineProcessingTypeCompat       EngineProcessingType = "compat"
	EngineProcessingTypeEscalation   EngineProcessingType = "escalation"
	EngineProcessingTypeHeartbeat    EngineProcessingType = "heartbeat"
	EngineProcessingTypeMessage      EngineProcessingType = "message"
	EngineProcessingTypeMetrics      EngineProcessingType = "metrics"
	EngineProcessingTypeNpCycle      EngineProcessingType = "np_cycle"
	EngineProcessingTypeRotation     EngineProcessingType = "rotation"
	EngineProcessingTypeSchedule     EngineProcessingType = "schedule"
	EngineProcessingTypeStatusUpdate EngineProcessingType = "status_update"
	EngineProcessingTypeVerify       EngineProcessingType = "verify"
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
	EnumAlertLogEventAcknowledged        EnumAlertLogEvent = "acknowledged"
	EnumAlertLogEventAssignmentChanged   EnumAlertLogEvent = "assignment_changed"
	EnumAlertLogEventClosed              EnumAlertLogEvent = "closed"
	EnumAlertLogEventCreated             EnumAlertLogEvent = "created"
	EnumAlertLogEventDuplicateSuppressed EnumAlertLogEvent = "duplicate_suppressed"
	EnumAlertLogEventEscalated           EnumAlertLogEvent = "escalated"
	EnumAlertLogEventEscalationRequest   EnumAlertLogEvent = "escalation_request"
	EnumAlertLogEventNoNotificationSent  EnumAlertLogEvent = "no_notification_sent"
	EnumAlertLogEventNotificationSent    EnumAlertLogEvent = "notification_sent"
	EnumAlertLogEventPolicyUpdated       EnumAlertLogEvent = "policy_updated"
	EnumAlertLogEventReopened            EnumAlertLogEvent = "reopened"
	EnumAlertLogEventResponseReceived    EnumAlertLogEvent = "response_received"
	EnumAlertLogEventStatusChanged       EnumAlertLogEvent = "status_changed"
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
	EnumAlertLogSubjectTypeChannel          EnumAlertLogSubjectType = "channel"
	EnumAlertLogSubjectTypeHeartbeatMonitor EnumAlertLogSubjectType = "heartbeat_monitor"
	EnumAlertLogSubjectTypeIntegrationKey   EnumAlertLogSubjectType = "integration_key"
	EnumAlertLogSubjectTypeUser             EnumAlertLogSubjectType = "user"
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
	EnumAlertSourceEmail                  EnumAlertSource = "email"
	EnumAlertSourceGeneric                EnumAlertSource = "generic"
	EnumAlertSourceGrafana                EnumAlertSource = "grafana"
	EnumAlertSourceManual                 EnumAlertSource = "manual"
	EnumAlertSourceNotify                 EnumAlertSource = "notify"
	EnumAlertSourcePrometheusAlertmanager EnumAlertSource = "prometheusAlertmanager"
	EnumAlertSourceSite24x7               EnumAlertSource = "site24x7"
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
	EnumAlertStatusActive    EnumAlertStatus = "active"
	EnumAlertStatusClosed    EnumAlertStatus = "closed"
	EnumAlertStatusTriggered EnumAlertStatus = "triggered"
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
	EnumHeartbeatStateHealthy   EnumHeartbeatState = "healthy"
	EnumHeartbeatStateInactive  EnumHeartbeatState = "inactive"
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
	EnumIntegrationKeysTypeEmail                  EnumIntegrationKeysType = "email"
	EnumIntegrationKeysTypeGeneric                EnumIntegrationKeysType = "generic"
	EnumIntegrationKeysTypeGrafana                EnumIntegrationKeysType = "grafana"
	EnumIntegrationKeysTypeNotify                 EnumIntegrationKeysType = "notify"
	EnumIntegrationKeysTypePrometheusAlertmanager EnumIntegrationKeysType = "prometheusAlertmanager"
	EnumIntegrationKeysTypeSite24x7               EnumIntegrationKeysType = "site24x7"
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
	EnumLimitTypeCalendarSubscriptionsPerUser EnumLimitType = "calendar_subscriptions_per_user"
	EnumLimitTypeContactMethodsPerUser        EnumLimitType = "contact_methods_per_user"
	EnumLimitTypeEpActionsPerStep             EnumLimitType = "ep_actions_per_step"
	EnumLimitTypeEpStepsPerPolicy             EnumLimitType = "ep_steps_per_policy"
	EnumLimitTypeHeartbeatMonitorsPerService  EnumLimitType = "heartbeat_monitors_per_service"
	EnumLimitTypeIntegrationKeysPerService    EnumLimitType = "integration_keys_per_service"
	EnumLimitTypeNotificationRulesPerUser     EnumLimitType = "notification_rules_per_user"
	EnumLimitTypeParticipantsPerRotation      EnumLimitType = "participants_per_rotation"
	EnumLimitTypeRulesPerSchedule             EnumLimitType = "rules_per_schedule"
	EnumLimitTypeTargetsPerSchedule           EnumLimitType = "targets_per_schedule"
	EnumLimitTypeUnackedAlertsPerService      EnumLimitType = "unacked_alerts_per_service"
	EnumLimitTypeUserOverridesPerSchedule     EnumLimitType = "user_overrides_per_schedule"
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
	EnumNotifChannelTypeSLACKUSERGROUP EnumNotifChannelType = "SLACK_USER_GROUP"
	EnumNotifChannelTypeWEBHOOK        EnumNotifChannelType = "WEBHOOK"
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
	EnumOutgoingMessagesStatusBundled        EnumOutgoingMessagesStatus = "bundled"
	EnumOutgoingMessagesStatusDelivered      EnumOutgoingMessagesStatus = "delivered"
	EnumOutgoingMessagesStatusFailed         EnumOutgoingMessagesStatus = "failed"
	EnumOutgoingMessagesStatusPending        EnumOutgoingMessagesStatus = "pending"
	EnumOutgoingMessagesStatusQueuedRemotely EnumOutgoingMessagesStatus = "queued_remotely"
	EnumOutgoingMessagesStatusSending        EnumOutgoingMessagesStatus = "sending"
	EnumOutgoingMessagesStatusSent           EnumOutgoingMessagesStatus = "sent"
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
	EnumOutgoingMessagesTypeAlertNotificationBundle    EnumOutgoingMessagesType = "alert_notification_bundle"
	EnumOutgoingMessagesTypeAlertStatusUpdate          EnumOutgoingMessagesType = "alert_status_update"
	EnumOutgoingMessagesTypeAlertStatusUpdateBundle    EnumOutgoingMessagesType = "alert_status_update_bundle"
	EnumOutgoingMessagesTypeScheduleOnCallNotification EnumOutgoingMessagesType = "schedule_on_call_notification"
	EnumOutgoingMessagesTypeTestNotification           EnumOutgoingMessagesType = "test_notification"
	EnumOutgoingMessagesTypeVerificationMessage        EnumOutgoingMessagesType = "verification_message"
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
	EnumRotationTypeDaily  EnumRotationType = "daily"
	EnumRotationTypeHourly EnumRotationType = "hourly"
	EnumRotationTypeWeekly EnumRotationType = "weekly"
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
	EnumUserContactMethodTypeEMAIL   EnumUserContactMethodType = "EMAIL"
	EnumUserContactMethodTypePUSH    EnumUserContactMethodType = "PUSH"
	EnumUserContactMethodTypeSLACKDM EnumUserContactMethodType = "SLACK_DM"
	EnumUserContactMethodTypeSMS     EnumUserContactMethodType = "SMS"
	EnumUserContactMethodTypeVOICE   EnumUserContactMethodType = "VOICE"
	EnumUserContactMethodTypeWEBHOOK EnumUserContactMethodType = "WEBHOOK"
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
	EnumUserRoleAdmin   EnumUserRole = "admin"
	EnumUserRoleUnknown EnumUserRole = "unknown"
	EnumUserRoleUser    EnumUserRole = "user"
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
	CreatedAt       time.Time
	DedupKey        sql.NullString
	Details         string
	EscalationLevel int32
	ID              int64
	LastEscalation  sql.NullTime
	LastProcessed   sql.NullTime
	ServiceID       uuid.NullUUID
	Source          EnumAlertSource
	Status          EnumAlertStatus
	Summary         string
}

type AlertFeedback struct {
	AlertID     int64
	NoiseReason string
}

type AlertLog struct {
	AlertID             sql.NullInt64
	Event               EnumAlertLogEvent
	ID                  int64
	Message             string
	Meta                pqtype.NullRawMessage
	SubChannelID        uuid.NullUUID
	SubClassifier       string
	SubHbMonitorID      uuid.NullUUID
	SubIntegrationKeyID uuid.NullUUID
	SubType             NullEnumAlertLogSubjectType
	SubUserID           uuid.NullUUID
	Timestamp           sql.NullTime
}

type AlertMetric struct {
	AlertID     int64
	ClosedAt    time.Time
	Escalated   bool
	ID          int64
	ServiceID   uuid.UUID
	TimeToAck   sql.NullInt64
	TimeToClose sql.NullInt64
}

type AlertStatusSubscription struct {
	AlertID         int64
	ChannelID       uuid.NullUUID
	ContactMethodID uuid.NullUUID
	ID              int64
	LastAlertStatus EnumAlertStatus
}

type AuthBasicUser struct {
	ID           int64
	PasswordHash string
	UserID       uuid.UUID
	Username     string
}

type AuthLinkRequest struct {
	CreatedAt  time.Time
	ExpiresAt  time.Time
	ID         uuid.UUID
	Metadata   json.RawMessage
	ProviderID string
	SubjectID  string
}

type AuthNonce struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

type AuthSubject struct {
	CmID       uuid.NullUUID
	ID         int64
	ProviderID string
	SubjectID  string
	UserID     uuid.UUID
}

type AuthUserSession struct {
	CreatedAt    time.Time
	ID           uuid.UUID
	LastAccessAt time.Time
	UserAgent    string
	UserID       uuid.NullUUID
}

type Config struct {
	CreatedAt time.Time
	Data      []byte
	ID        int32
	Schema    int32
}

type ConfigLimit struct {
	ID  EnumLimitType
	Max int32
}

type EngineProcessingVersion struct {
	State   json.RawMessage
	TypeID  EngineProcessingType
	Version int32
}

type EpStepOnCallUser struct {
	EndTime   sql.NullTime
	EpStepID  uuid.UUID
	ID        int64
	StartTime time.Time
	UserID    uuid.UUID
}

type EscalationPolicy struct {
	Description string
	ID          uuid.UUID
	Name        string
	Repeat      int32
	StepCount   int32
}

type EscalationPolicyAction struct {
	ChannelID              uuid.NullUUID
	EscalationPolicyStepID uuid.UUID
	ID                     uuid.UUID
	RotationID             uuid.NullUUID
	ScheduleID             uuid.NullUUID
	UserID                 uuid.NullUUID
}

type EscalationPolicyState struct {
	AlertID                    int64
	EscalationPolicyID         uuid.UUID
	EscalationPolicyStepID     uuid.NullUUID
	EscalationPolicyStepNumber int32
	ForceEscalation            bool
	ID                         int64
	LastEscalation             sql.NullTime
	LoopCount                  int32
	NextEscalation             sql.NullTime
	ServiceID                  uuid.UUID
}

type EscalationPolicyStep struct {
	Delay              int32
	EscalationPolicyID uuid.UUID
	ID                 uuid.UUID
	StepNumber         int32
}

type GorpMigration struct {
	AppliedAt sql.NullTime
	ID        string
}

type HeartbeatMonitor struct {
	HeartbeatInterval int64
	ID                uuid.UUID
	LastHeartbeat     sql.NullTime
	LastState         EnumHeartbeatState
	Name              string
	ServiceID         uuid.UUID
}

type IntegrationKey struct {
	ID        uuid.UUID
	Name      string
	ServiceID uuid.UUID
	Type      EnumIntegrationKeysType
}

type Keyring struct {
	ID               string
	NextKey          []byte
	NextRotation     sql.NullTime
	RotationCount    int64
	SigningKey       []byte
	VerificationKeys []byte
}

type Label struct {
	ID           int64
	Key          string
	TgtServiceID uuid.UUID
	Value        string
}

type NotificationChannel struct {
	CreatedAt time.Time
	ID        uuid.UUID
	Meta      json.RawMessage
	Name      string
	Type      EnumNotifChannelType
	Value     string
}

type NotificationPolicyCycle struct {
	AlertID     int32
	Checked     bool
	ID          uuid.UUID
	LastTick    sql.NullTime
	RepeatCount int32
	StartedAt   time.Time
	UserID      uuid.UUID
}

type OutgoingMessage struct {
	AlertID                sql.NullInt64
	AlertLogID             sql.NullInt64
	ChannelID              uuid.NullUUID
	ContactMethodID        uuid.NullUUID
	CreatedAt              time.Time
	CycleID                uuid.NullUUID
	EscalationPolicyID     uuid.NullUUID
	FiredAt                sql.NullTime
	ID                     uuid.UUID
	LastStatus             EnumOutgoingMessagesStatus
	LastStatusAt           sql.NullTime
	MessageType            EnumOutgoingMessagesType
	NextRetryAt            sql.NullTime
	ProviderMsgID          sql.NullString
	ProviderSeq            int32
	RetryCount             int32
	ScheduleID             uuid.NullUUID
	SendingDeadline        sql.NullTime
	SentAt                 sql.NullTime
	ServiceID              uuid.NullUUID
	SrcValue               sql.NullString
	StatusAlertIds         []int64
	StatusDetails          string
	UserID                 uuid.NullUUID
	UserVerificationCodeID uuid.NullUUID
}

type RegionID struct {
	ID   int32
	Name string
}

type Rotation struct {
	Description      string
	ID               uuid.UUID
	LastProcessed    sql.NullTime
	Name             string
	ParticipantCount int32
	ShiftLength      int64
	StartTime        time.Time
	TimeZone         string
	Type             EnumRotationType
}

type RotationParticipant struct {
	ID         uuid.UUID
	Position   int32
	RotationID uuid.UUID
	UserID     uuid.UUID
}

type RotationState struct {
	ID                    int64
	Position              int32
	RotationID            uuid.UUID
	RotationParticipantID uuid.UUID
	ShiftStart            time.Time
	Version               int32
}

type Schedule struct {
	Description   string
	ID            uuid.UUID
	LastProcessed sql.NullTime
	Name          string
	TimeZone      string
}

type ScheduleDatum struct {
	Data          json.RawMessage
	ID            int64
	LastCleanupAt sql.NullTime
	ScheduleID    uuid.UUID
}

type ScheduleOnCallUser struct {
	EndTime    sql.NullTime
	ID         int64
	ScheduleID uuid.UUID
	StartTime  time.Time
	UserID     uuid.UUID
}

type ScheduleRule struct {
	CreatedAt     time.Time
	EndTime       time.Time
	Friday        bool
	ID            uuid.UUID
	IsActive      bool
	Monday        bool
	Saturday      bool
	ScheduleID    uuid.UUID
	StartTime     time.Time
	Sunday        bool
	TgtRotationID uuid.NullUUID
	TgtUserID     uuid.NullUUID
	Thursday      bool
	Tuesday       bool
	Wednesday     bool
}

type Service struct {
	Description          string
	EscalationPolicyID   uuid.UUID
	ID                   uuid.UUID
	MaintenanceExpiresAt sql.NullTime
	Name                 string
}

type SwitchoverLog struct {
	Data      json.RawMessage
	ID        int64
	Timestamp time.Time
}

type SwitchoverState struct {
	CurrentState EnumSwitchoverState
	DbID         uuid.UUID
	Ok           bool
}

type TwilioSmsCallback struct {
	AlertID     sql.NullInt64
	CallbackID  uuid.UUID
	Code        int32
	ID          int64
	PhoneNumber string
	SentAt      time.Time
	ServiceID   uuid.NullUUID
}

type TwilioSmsError struct {
	ErrorMessage string
	ID           int64
	OccurredAt   time.Time
	Outgoing     bool
	PhoneNumber  string
}

type TwilioVoiceError struct {
	ErrorMessage string
	ID           int64
	OccurredAt   time.Time
	Outgoing     bool
	PhoneNumber  string
}

type User struct {
	AlertStatusLogContactMethodID uuid.NullUUID
	AvatarUrl                     string
	Bio                           string
	Email                         string
	ID                            uuid.UUID
	Name                          string
	Role                          EnumUserRole
}

type UserCalendarSubscription struct {
	Config     json.RawMessage
	CreatedAt  time.Time
	Disabled   bool
	ID         uuid.UUID
	LastAccess sql.NullTime
	LastUpdate time.Time
	Name       string
	ScheduleID uuid.UUID
	UserID     uuid.UUID
}

type UserContactMethod struct {
	Disabled            bool
	EnableStatusUpdates bool
	ID                  uuid.UUID
	LastTestVerifyAt    sql.NullTime
	Metadata            pqtype.NullRawMessage
	Name                string
	Pending             bool
	Type                EnumUserContactMethodType
	UserID              uuid.UUID
	Value               string
}

type UserFavorite struct {
	ID                    int64
	TgtEscalationPolicyID uuid.NullUUID
	TgtRotationID         uuid.NullUUID
	TgtScheduleID         uuid.NullUUID
	TgtServiceID          uuid.NullUUID
	TgtUserID             uuid.NullUUID
	UserID                uuid.UUID
}

type UserNotificationRule struct {
	ContactMethodID uuid.UUID
	CreatedAt       sql.NullTime
	DelayMinutes    int32
	ID              uuid.UUID
	UserID          uuid.UUID
}

type UserOverride struct {
	AddUserID     uuid.NullUUID
	EndTime       time.Time
	ID            uuid.UUID
	RemoveUserID  uuid.NullUUID
	StartTime     time.Time
	TgtScheduleID uuid.UUID
}

type UserSlackDatum struct {
	AccessToken string
	ID          uuid.UUID
}

type UserVerificationCode struct {
	Code            int32
	ContactMethodID uuid.UUID
	ExpiresAt       time.Time
	ID              uuid.UUID
	Sent            bool
}
