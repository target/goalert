package nfynet

import (
	"fmt"
	"strings"
)

// TargetID is a notification target.
type TargetID struct {
	NetworkID NetworkID
	ID        string
}

// String will return the string representation of the target.
func (t TargetID) String() string {
	return fmt.Sprintf("T|%s|%s|%s|%s",
		escape(t.NetworkID.ID),
		escape(t.NetworkID.SubNetID),
		escape(t.NetworkID.SubTypeID),
		escape(t.ID),
	)
}

// ParseTargetID will parse a string representation of a target.
func ParseTargetID(s string) (*TargetID, error) {
	err := validateEncoding(s)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(s, "|")
	if parts[0] != "T" {
		// could be newer version
		return nil, fmt.Errorf("unknown target string: %s", s)
	}
	if len(parts) != 5 {
		return nil, fmt.Errorf("invalid target string: %s", s)
	}

	return &TargetID{
		NetworkID: NetworkID{
			ID:        unescape(parts[1]),
			SubNetID:  unescape(parts[2]),
			SubTypeID: unescape(parts[3]),
		},
		ID: unescape(parts[4]),
	}, nil
}
