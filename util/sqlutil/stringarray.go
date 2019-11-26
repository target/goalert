package sqlutil

import (
	"database/sql/driver"

	"github.com/jackc/pgx/pgtype"
)

type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	var pgArray pgtype.TextArray
	err := pgArray.Set([]string(s))
	if err != nil {
		return nil, err
	}
	return pgArray.Value()
}
func (s *StringArray) Scan(src interface{}) error {
	var pgArray pgtype.TextArray
	err := pgArray.Scan(src)
	if err != nil {
		return err
	}

	return pgArray.AssignTo((*[]string)(s))
}
