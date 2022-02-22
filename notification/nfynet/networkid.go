package nfynet

import (
	"fmt"
	"strings"
)

// NetworkID is a network of notification targets.
type NetworkID struct {

	// ID is the unique identifier of the network.
	ID string

	// SubNetID is an optional identifier of the sub-network (e.g., Slack workspace/team ID)
	SubNetID string

	// SubTypeID is an optional identifier of the sub-type (e.g., "channel", "user", "sms", "voice", etc.)
	SubTypeID string
}

// String will return the string representation of the network.
func (n NetworkID) String() string {
	return fmt.Sprintf("N|%s|%s|%s",
		escape(n.ID),
		escape(n.SubNetID),
		escape(n.SubTypeID),
	)
}

// ParseNetworkID will parse a string representation of a network.
func ParseNetworkID(s string) (*NetworkID, error) {
	err := validateEncoding(s)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(s, "|")
	if parts[0] != "N" {
		// could be newer version
		return nil, fmt.Errorf("unknown network string: %s", s)
	}
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid network string: %s", s)
	}

	return &NetworkID{
		ID:        unescape(parts[1]),
		SubNetID:  unescape(parts[2]),
		SubTypeID: unescape(parts[3]),
	}, nil
}
