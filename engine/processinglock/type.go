package processinglock

import (
	"database/sql/driver"
	"fmt"
	"github.com/target/goalert/validation/validate"
)

// Type indicates the lock type. For TypeMessage, the RegionID is used.
type Type string

// Recognized types
const (
	TypeEscalation   Type = "escalation"
	TypeHeartbeat    Type = "heartbeat"
	TypeNPCycle      Type = "np_cycle"
	TypeRotation     Type = "rotation"
	TypeSchedule     Type = "schedule"
	TypeStatusUpdate Type = "status_update"
	TypeVerify       Type = "verify"
	TypeMessage      Type = "message"
	TypeCleanup      Type = "cleanup"
)

func (t Type) validate() error {
	return validate.OneOf("Type", t,
		TypeEscalation,
		TypeHeartbeat,
		TypeNPCycle,
		TypeRotation,
		TypeSchedule,
		TypeStatusUpdate,
		TypeVerify,
		TypeMessage,
		TypeCleanup,
	)
}

// Value will return the DB enum value of the Type.
func (t Type) Value() (driver.Value, error) {
	return string(t), t.validate()
}

// Scan will scan a DB enum value into Type.
func (t *Type) Scan(value interface{}) error {
	switch _t := value.(type) {
	case []byte:
		*t = Type(_t)
	case string:
		*t = Type(_t)
	default:
		return fmt.Errorf("could not process unknown type for Type(%T)", t)
	}
	return t.validate()
}

// LockID returns the int value used for the advisory lock for the Type.
func (t Type) LockID() int {
	switch t {
	case TypeEscalation:
		return 0x1000 // 4096
	case TypeHeartbeat:
		return 0x1010 // 4112
	case TypeNPCycle:
		return 0x1020 // 4128
	case TypeRotation:
		return 0x1030 // 4144
	case TypeSchedule:
		return 0x1040 // 4160
	case TypeStatusUpdate:
		return 0x1050 // 4176
	case TypeVerify:
		return 0x1060 // 4192
	case TypeMessage:
		return 0x1070 // 4208
	case TypeCleanup:
		return 0x1080 // 4224
	}

	panic("invalid type")
}
