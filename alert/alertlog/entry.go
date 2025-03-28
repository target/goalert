package alertlog

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/util/log"
)

type Entry struct {
	id        int
	alertID   int
	timestamp time.Time
	_type     Type
	message   string
	subject   struct {
		_type                SubjectType
		userID               uuid.NullUUID
		userName             sql.NullString
		integrationKeyID     uuid.NullUUID
		integrationKeyName   sql.NullString
		heartbeatMonitorID   uuid.NullUUID
		heartbeatMonitorName sql.NullString
		channelID            uuid.NullUUID
		channelName          sql.NullString
		classifier           string
	}
	meta rawJSON
}

func (e Entry) Meta(ctx context.Context) interface{} {
	var dest interface{}
	switch e.Type() {
	case TypeEscalated:
		dest = &EscalationMetaData{}
	case TypeNotificationSent:
		dest = &NotificationMetaData{}
	case TypeCreated:
		dest = &CreatedMetaData{}
	case TypeClosed:
		dest = &AutoClose{}
	default:
		return nil
	}

	err := json.Unmarshal(e.meta, dest)
	if err != nil {
		log.Debug(ctx, err)
		return nil
	}
	return dest
}

func (e Entry) AlertID() int {
	return e.alertID
}

func (e Entry) MessageID(ctx context.Context) string {
	m := e.Meta(ctx)
	if m == nil {
		return ""
	}
	if m, ok := m.(*NotificationMetaData); ok && m != nil {
		return m.MessageID
	}
	return ""
}

func (e Entry) ID() int {
	return e.id
}

func (e Entry) Timestamp() time.Time {
	return e.timestamp
}

func (e Entry) Type() Type {
	switch e._type {
	case _TypeResponseReceived:
		return respRecvType(e.message)
	case _TypeStatusChanged:
		return statChgType(e.message)
	}

	return e._type
}

func (e Entry) Subject() *Subject {
	if e.subject._type == SubjectTypeNone {
		if e.message != "" {
			return e.subjectFromMessage()
		}
		return nil
	}

	s := &Subject{
		Type:       e.subject._type,
		Classifier: e.subject.classifier,
	}

	switch s.Type {
	case SubjectTypeUser:
		s.ID = e.subject.userID.UUID.String()
		s.Name = e.subject.userName.String
	case SubjectTypeIntegrationKey:
		s.ID = e.subject.integrationKeyID.UUID.String()
		s.Name = e.subject.integrationKeyName.String
	case SubjectTypeHeartbeatMonitor:
		s.ID = e.subject.heartbeatMonitorID.UUID.String()
		s.Name = e.subject.heartbeatMonitorName.String
	case SubjectTypeChannel:
		s.ID = e.subject.channelID.UUID.String()
		s.Name = e.subject.channelName.String
	}

	return s
}

func escalationMsg(m *EscalationMetaData) string {
	msg := fmt.Sprintf(" to step #%d", m.NewStepIndex+1)
	if m.Repeat {
		msg += " (policy repeat)"
	}
	if m.Forced {
		msg += " due to manual escalation"
	} else if m.Deleted {
		msg += " due to current step being deleted"
	} else if m.OldDelayMinutes > 0 {
		msg += fmt.Sprintf(" automatically after %d minutes", m.OldDelayMinutes)
	}

	return msg
}

func (e Entry) String(ctx context.Context) string {
	var msg string
	var infinitive bool
	switch e.Type() {
	case TypeCreated:
		msg = "Created"
	case TypeAcknowledged:
		msg = "Acknowledged"
	case TypeClosed:
		msg = "Closed"
		meta, ok := e.Meta(ctx).(*AutoClose)
		if ok {
			msg = "Closed due to inactivity (unacknowledged for  " + strconv.Itoa(meta.AlertAutoCloseDays) + " days)"
		}

	case TypeEscalated:
		msg = "Escalated"
		meta, ok := e.Meta(ctx).(*EscalationMetaData)
		if ok {
			msg += escalationMsg(meta)
		}
	case TypeNotificationSent:
		msg = "Notification sent"
		infinitive = true
	case TypeNoNotificationSent:
		msg = "No notification sent"
		infinitive = true
	case TypePolicyUpdated:
		msg = "Policy updated"
	case TypeDuplicateSupressed:
		msg = "Suppressed duplicate: created"
	case TypeEscalationRequest:
		msg = "Escalation requested"
	default:
		return "Error"
	}

	// include subject, if available
	msg += subjectString(infinitive, e.Subject())

	return msg
}

func (e *Entry) scanWith(scan func(...interface{}) error) error {
	return scan(
		&e.id,
		&e.alertID,
		&e.timestamp,
		&e._type,
		&e.message,
		&e.subject._type,
		&e.subject.userID,
		&e.subject.userName,
		&e.subject.integrationKeyID,
		&e.subject.integrationKeyName,
		&e.subject.heartbeatMonitorID,
		&e.subject.heartbeatMonitorName,
		&e.subject.channelID,
		&e.subject.channelName,
		&e.subject.classifier,
		&e.meta,
	)
}
