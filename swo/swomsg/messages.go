package swomsg

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID   uuid.UUID
	Node uuid.UUID
	TS   time.Time `json:"-"`

	Type  string
	Ack   bool            `json:",omitempty"`
	AckID uuid.UUID       `json:",omitempty"`
	Data  json.RawMessage `json:",omitempty"`
}
