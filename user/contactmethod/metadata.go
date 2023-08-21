package contactmethod

import (
	"encoding/json"
	"fmt"
	"time"
)

// Metadata stores information about a contact method.
type Metadata struct {
	FetchedAt time.Time `json:"-"`

	CarrierV1 struct {
		UpdatedAt         time.Time
		Name              string
		Type              string
		MobileNetworkCode string
		MobileCountryCode string
	}
}

// MarshalJSON implements `json.Marshaler`. It is used to allow `omitempty` behavior
// with embedded structs.
func (m Metadata) MarshalJSON() ([]byte, error) {
	var enc struct {
		CarrierV1 json.RawMessage `json:",omitempty"`
	}

	if !m.CarrierV1.UpdatedAt.IsZero() {
		data, err := json.Marshal(m.CarrierV1)
		if err != nil {
			return nil, fmt.Errorf("marshal CarrierV1: %w", err)
		}
		enc.CarrierV1 = data
	}

	return json.Marshal(enc)
}
