package alertlog

import (
	"regexp"
	"strings"
)

func (e *rawEntry) subjectFromMessage() *Subject {
	switch e._type {
	case TypeCreated:
		return createdSubject(e.message)
	case TypeNotificationSent:
		return notifSentSubject(e.message)
	case _TypeResponseReceived:
		return respRecvSubject(e.message)
	case _TypeStatusChanged:
		return statChgSubject(e.message)
	}

	return nil
}

var (
	respRx  = regexp.MustCompile(`^(Closed|Acknowledged) by (.*) via (SMS|VOICE)$`)
	notifRx = regexp.MustCompile(`^Notification sent to (.*) via (SMS|VOICE)$`)
	statRx  = regexp.MustCompile(`^status changed to (active|closed) by (.*)( via UI)?$`)
)

func statChgSubject(msg string) *Subject {
	parts := statRx.FindStringSubmatch(msg)
	if len(parts) == 0 {
		return nil
	}
	return &Subject{
		Type: SubjectTypeUser,
		Name: parts[2],
	}
}

func statChgType(msg string) Type {
	if msg == "Status updated from active to triggered" {
		return TypeEscalated
	} else if msg == "Status updated from triggered to active" {
		return TypeAcknowledged
	} else if strings.HasPrefix(msg, "status changed to closed") {
		return TypeClosed
	} else if strings.HasPrefix(msg, "status changed to active") {
		return TypeAcknowledged
	}
	return ""
}

func respRecvType(msg string) Type {
	parts := respRx.FindStringSubmatch(msg)
	if len(parts) == 0 {
		return ""
	}
	switch parts[1] {
	case "Closed":
		return TypeClosed
	case "Acknowledged":
		return TypeAcknowledged
	}
	return ""
}
func respRecvSubject(msg string) *Subject {
	parts := respRx.FindStringSubmatch(msg)
	if len(parts) == 0 {
		return nil
	}
	if parts[3] == "VOICE" {
		parts[3] = "Voice"
	}
	return &Subject{
		Type:       SubjectTypeUser,
		Name:       parts[2],
		Classifier: parts[3],
	}
}

func notifSentSubject(msg string) *Subject {
	parts := notifRx.FindStringSubmatch(msg)
	if len(parts) == 0 {
		return nil
	}

	if parts[2] == "VOICE" {
		parts[2] = "Voice"
	}

	return &Subject{
		Type:       SubjectTypeUser,
		Name:       parts[1],
		Classifier: parts[2],
	}
}

func createdSubject(msg string) *Subject {
	switch msg {
	case "Created via: grafana":
		return &Subject{Type: SubjectTypeIntegrationKey, Classifier: "Grafana"}
	case "Created via: manual":
		return &Subject{Type: SubjectTypeUser, Classifier: "Web"}
	case "Created via: generic":
		return &Subject{Type: SubjectTypeIntegrationKey, Classifier: "Generic"}
	}
	return nil
}
