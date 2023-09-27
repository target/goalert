package signal

import "time"

type Signal struct {
	ID              int64                  `json:"_id"`
	ServiceID       string                 `json:"service_id"`
	ServiceRuleID   string                 `json:"service_rule_id"`
	OutgoingPayload map[string]interface{} `json:"outgoing_payload"`
	Scheduled       bool                   `json:"scheduled"`
	Timestamp       time.Time              `json:"timestamp"`
}
