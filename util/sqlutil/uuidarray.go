package sqlutil

import (
	"database/sql/driver"

	"github.com/jackc/pgtype"
)

type NullUUIDArray struct {
	UUIDArray UUIDArray
	Valid     bool
}

func (s NullUUIDArray) Value() (driver.Value, error) {
	if !s.Valid {
		return nil, nil
	}

	return s.UUIDArray.Value()
}
func (s *NullUUIDArray) Scan(src interface{}) error {
	switch src.(type) {
	case nil:
		s.Valid = false
		s.UUIDArray = nil
		return nil
	default:
	}

	err := s.UUIDArray.Scan(src)
	s.Valid = err == nil
	return err
}

type UUIDArray []string

func (s UUIDArray) Value() (driver.Value, error) {
	var pgArray pgtype.UUIDArray
	err := pgArray.Set([]string(s))
	if err != nil {
		return nil, err
	}
	return pgArray.Value()
}
func (s *UUIDArray) Scan(src interface{}) error {
	var pgArray pgtype.UUIDArray
	err := pgArray.Scan(src)
	if err != nil {
		return err
	}

	return pgArray.AssignTo((*[]string)(s))
}
