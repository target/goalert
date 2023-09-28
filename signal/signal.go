package signal

import (
	"encoding/json"
	"time"
)

type Signal struct {
	ID              int64                  `json:"_id"`
	ServiceID       string                 `json:"service_id"`
	ServiceRuleID   string                 `json:"service_rule_id"`
	OutgoingPayload map[string]interface{} `json:"outgoing_payload"`
	Scheduled       bool                   `json:"scheduled"`
	Timestamp       time.Time              `json:"timestamp"`
}

func (s *Signal) scanFrom(scanFn func(...interface{}) error) error {
	var outgoingPayloadBytes []byte
	err := scanFn(&s.ID, &s.ServiceRuleID, &s.ServiceID, &outgoingPayloadBytes, &s.Scheduled, &s.Timestamp)
	if err != nil {
		return err
	}
	return json.Unmarshal(outgoingPayloadBytes, &s.OutgoingPayload)
}
