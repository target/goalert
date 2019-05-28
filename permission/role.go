package permission

import (
	"database/sql/driver"
	"fmt"
)

// Role represents a users access level
type Role string

// Available roles
const (
	RoleUser    Role = "user"
	RoleAdmin   Role = "admin"
	RoleUnknown Role = "unknown"
)

// Scan handles reading a Role from the DB format
func (r *Role) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		*r = Role(t)
	case string:
		*r = Role(t)
	default:
		return fmt.Errorf("could not process unknown type for role %T", t)
	}

	if *r != RoleAdmin && *r != RoleUser && *r != RoleUnknown {
		return fmt.Errorf("unknown value for role %v", *r)
	}

	return nil
}

// Value converts the Role to the DB representation
func (r Role) Value() (driver.Value, error) {
	switch r {
	case RoleUser, RoleAdmin, RoleUnknown:
		return string(r), nil
	default:
		return nil, fmt.Errorf("invalid role value: %v", r)
	}
}
