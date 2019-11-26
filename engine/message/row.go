package message

import (
	"database/sql"
	"github.com/target/goalert/notification"
	"time"
)

type batchCounts map[notification.DestType]int

type row struct {
	ID         string
	CreatedAt  time.Time
	Type       Type
	DestType   notification.DestType
	DestID     string
	AlertID    sql.NullInt64
	AlertLogID sql.NullInt64
	VerifyID   sql.NullString
}
