package sqlutil

import (
	"database/sql/driver"
	"time"
)

type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not null
}

func (n NullTime) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}

	return n.Time, nil
}

func (n *NullTime) Scan(src interface{}) error {
	n.Time, n.Valid = src.(time.Time)
	return nil
}
