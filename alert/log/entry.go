package alertlog

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

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
		userID               sql.NullString
		userName             sql.NullString
		integrationKeyID     sql.NullString
		integrationKeyName   sql.NullString
		heartbeatMonitorID   sql.NullString
		heartbeatMonitorName sql.NullString
		channelID            sql.NullString
		channelName          sql.NullString
		classifier           string
	}
	meta          rawJSON
	lastStatus    sql.NullString
	statusDetails sql.NullString
}

func (e Entry) Meta() interface{} {
	switch e.Type() {
	case TypeEscalated:
		var esc EscalationMetaData
		err := json.Unmarshal(e.meta, &esc)
		if err != nil {
			log.Debug(context.Background(), err)
			return nil
		}
		return &esc
	}

	return nil
}
func (e Entry) AlertID() int {
	return e.alertID
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
		s.ID = e.subject.userID.String
		s.Name = e.subject.userName.String
	case SubjectTypeIntegrationKey:
		s.ID = e.subject.integrationKeyID.String
		s.Name = e.subject.integrationKeyName.String
	case SubjectTypeHeartbeatMonitor:
		s.ID = e.subject.heartbeatMonitorID.String
		s.Name = e.subject.heartbeatMonitorName.String
	case SubjectTypeChannel:
		s.ID = e.subject.channelID.String
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

func (e Entry) String() string {
	var msg string
	var infinitive bool
	switch e.Type() {
	case TypeCreated:
		msg = "Created"
	case TypeAcknowledged:
		msg = "Acknowledged"
	case TypeClosed:
		msg = "Closed"
	case TypeEscalated:
		msg = "Escalated"
		meta, ok := e.Meta().(*EscalationMetaData)
		if ok {
			msg += escalationMsg(meta)
		}
	case TypeNotificationSent:
		msg = "Notification sent"
		infinitive = true
	case TypeNoNotificationSent:
		msg = "No notification sent"
		infinitive = true
	case TypeNotificationSendFailure:
		msg = "Notification failed to send"
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
		&e.lastStatus,
		&e.statusDetails,
	)
}
