package sqlutil

import (
	"database/sql/driver"

	"github.com/jackc/pgx/pgtype"
)

type BoolArray []bool

func (s BoolArray) Value() (driver.Value, error) {
	var pgArray pgtype.BoolArray
	err := pgArray.Set([]bool(s))
	if err != nil {
		return nil, err
	}
	return pgArray.Value()
}
func (s *BoolArray) Scan(src interface{}) error {
	var pgArray pgtype.BoolArray
	err := pgArray.Scan(src)
	if err != nil {
		return err
	}

	return pgArray.AssignTo((*[]bool)(s))
}
