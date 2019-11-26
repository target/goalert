package sqlutil

import (
	"database/sql/driver"

	"github.com/jackc/pgx/pgtype"
)

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
