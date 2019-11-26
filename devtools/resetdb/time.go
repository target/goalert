package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"

	"github.com/jackc/pgx/pgtype"
	"github.com/target/goalert/schedule/rule"
)

// pgTime handles encoding rule.Clock values with the native (binary) pgx interface.
type pgTime rule.Clock

func (t pgTime) AssignTo(dst interface{}) error { return errors.New("cannot AssignTo") }
func (t pgTime) Get() interface{}               { return time.Duration(t).Seconds() }
func (t *pgTime) Set(src interface{}) error {
	switch val := src.(type) {
	case rule.Clock:
		*t = pgTime(val)
	case *rule.Clock:
		*t = pgTime(*val)
	default:
		return errors.New("invalid type")
	}
	return nil
}

func (t pgTime) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	b := bytes.NewBuffer(buf)
	err := binary.Write(b, binary.BigEndian, time.Duration(t).Truncate(time.Minute).Microseconds())
	return b.Bytes(), err
}
