package sqlutil

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/pgtype"
)

type IntArray []int64

func (s IntArray) Value() (driver.Value, error) {
	var pgArray pgtype.Int8Array

	err := pgArray.Set([]int64(s))
	if err != nil {
		return nil, err
	}

	fmt.Println("SET", s)
	return pgArray.Value()
}

func (s *IntArray) Scan(src interface{}) error {
	var pgArray pgtype.Int8Array

	err := pgArray.Scan(src)
	if err != nil {
		return err
	}

	return pgArray.AssignTo((*[]int64)(s))
}
