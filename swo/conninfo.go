package swo

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// ConnInfo contains information about a connection to the DB for SWO.
//
// This is stored as the `application_name` for a connection in Postgres in
// the format of "GoAlert <version> SWO:<type>:<id>" where id is a base64
// encoded UUID that should match what ends up in a `switchover_log` hello message.
type ConnInfo struct {
	Version string
	Type    ConnType
	ID      uuid.UUID
}

// ConnType indicates a type of SWO connection.
type ConnType byte

const (
	// ConnTypeMainMgr is the connection pool to the main/old DB used to coordinate the switchover.
	ConnTypeMainMgr ConnType = iota + 'A'

	// ConnTypeMainApp is the connection pool used by the GoAlert application to the main/old DB.
	//
	// Connections here are protected with a shared advisory lock.
	ConnTypeMainApp

	// ConnTypeNextMgr is the connection pool to the next/new DB used for applying changes during the switchover.
	ConnTypeNextMgr

	// ConnTypeNextApp is the connection pool used by the GoAlert application to the next/new DB, after the switchover is completed.
	ConnTypeNextApp
)

// IsNext returns true if the connection is for the next DB.
func (t ConnType) IsNext() bool {
	return t == ConnTypeNextMgr || t == ConnTypeNextApp
}

// IsValid returns true if the ConnType is valid.
func (t ConnType) IsValid() bool {
	return t >= ConnTypeMainMgr && t <= ConnTypeNextApp
}

// String returns a string representation of the ConnInfo.
func (c ConnInfo) String() string {
	// ensure c.Version is <= 24 characters
	if len(c.Version) > 24 {
		c.Version = c.Version[:24]
	}

	if !c.Type.IsValid() {
		panic(fmt.Sprintf("invalid connection type: 0x%0x", c.Type))
	}

	id := base64.RawURLEncoding.EncodeToString(c.ID[:])
	return fmt.Sprintf("GoAlert %s SWO:%c:%s", c.Version, c.Type, id)
}

// ParseConnInfo parses a connection string into a ConnInfo.
func ParseConnInfo(s string) (*ConnInfo, error) {
	if !strings.HasPrefix(s, "GoAlert ") {
		return nil, fmt.Errorf("missing 'GoAlert' prefix: %q", s)
	}
	s = strings.TrimPrefix(s, "GoAlert ")

	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return nil, fmt.Errorf("incorrect number of segments: %q", s)
	}

	if !strings.HasSuffix(parts[0], " SWO") {
		return nil, fmt.Errorf("missing 'SWO' suffix: %q", s)
	}
	parts[0] = strings.TrimSuffix(parts[0], " SWO")

	var info ConnInfo
	info.Version = parts[0]

	if len(parts[1]) != 1 {
		return nil, fmt.Errorf("invalid connection type: %q", s)
	}
	info.Type = ConnType(parts[1][0])
	if !info.Type.IsValid() {
		return nil, fmt.Errorf("invalid connection type: %q", s)
	}

	id, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid ID '%s': %w", parts[2], err)
	}
	if len(id) != 16 {
		return nil, fmt.Errorf("invalid ID '%s': incorrect length", parts[2])
	}
	copy(info.ID[:], id)

	return &info, nil
}
