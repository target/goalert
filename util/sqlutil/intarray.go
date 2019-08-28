package sqlutil

import (
	"database/sql/driver"
	"github.com/jackc/pgx/pgtype"
)

type IntArray []int

func (s IntArray) Value() (driver.Value, error) {
	var pgArray pgtype.Int8Array

	err := pgArray.Set([]int(s))
	if err != nil {
		return nil, err
	}

	return pgArray.Value()
}

func (s *IntArray) Scan(src interface{}) error {
	var pgArray pgtype.Int8Array

	err := pgArray.Scan(src)
	if err != nil {
		return err
	}

	return pgArray.AssignTo((*[]int)(s))
}
