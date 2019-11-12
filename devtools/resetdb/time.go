package main

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/jackc/pgtype"
	"github.com/target/goalert/schedule/rule"
)

// pgTime handles encoding rule.Clock values with the native (binary) pgx interface.
type pgTime rule.Clock

var _ pgtype.BinaryEncoder = pgTime(0)

func (t pgTime) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	b := bytes.NewBuffer(buf)
	err := binary.Write(b, binary.BigEndian, time.Duration(t).Truncate(time.Minute).Microseconds())
	return b.Bytes(), err
}
