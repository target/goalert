package swomsg

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Message represents a single event in the switchover log.
type Message struct {
	ID   uuid.UUID
	Node uuid.UUID
	TS   time.Time `json:"-"`

	Type  string
	Ack   bool            `json:",omitempty"`
	AckID uuid.UUID       `json:",omitempty"`
	Data  json.RawMessage `json:",omitempty"`
}
