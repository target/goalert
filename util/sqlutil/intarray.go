package sqlutil

import (
	"database/sql/driver"
	"github.com/jackc/pgx/pgtype"
)

type IntArray []int

func (a IntArray) Value() (driver.Value, error) {
	arr := make([]int64, len(a))
	for i, e := range a {
		arr[i] = int64(e)
	}

	var pgArray pgtype.Int8Array
	err := pgArray.Set(arr)
	if err != nil {
		return nil, err
	}

	return pgArray.Value()
}

func (a *IntArray) Scan(src interface{}) error {
	var pgArray pgtype.Int8Array

	err := pgArray.Scan(src)
	if err != nil {
		return err
	}

	var arr []int64
	err = pgArray.AssignTo(&arr)
	if err != nil {
		return err
	}

	*a = make(IntArray, len(arr))
	for i, e := range arr {
		(*a)[i] = int(e)
	}

	return nil
}
