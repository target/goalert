package sqlutil

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
)

type NullUUID struct {
	UUID  uuid.UUID
	Valid bool
}

func (u NullUUID) Value() (driver.Value, error) {
	if !u.Valid {
		return nil, nil
	}

	return []byte(u.UUID[:]), nil
}

func (u *NullUUID) Scan(src interface{}) (err error) {
	if src == nil {
		u.Valid = false
		u.UUID = uuid.UUID{}
		return nil
	}

	switch s := src.(type) {
	case string:
		u.UUID, err = uuid.Parse(s)
	case []byte:
		if len(s) == 16 {
			u.UUID, err = uuid.FromBytes(s)
			break
		}
		u.UUID, err = uuid.Parse(string(s))
	default:
		return fmt.Errorf("unknown format for UUID: %T", s)
	}
	u.Valid = err == nil

	return err
}
