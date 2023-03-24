package alert

import (
	"context"

	"github.com/target/goalert/alert/alertlog"

	"github.com/pkg/errors"
)

type LogEntryFetcher interface {
	// LogEntry fetchs the latest log entry for a given alertID and type.
	LogEntry(ctx context.Context) (*alertlog.Entry, error)
}

type logError struct {
	isAlreadyAcknowledged bool
	isAlreadyClosed       bool
	alertID               int
	_type                 alertlog.Type
	logDB                 *alertlog.Store
}

func (logError) ClientError() bool { return true }

func (e logError) LogEntry(ctx context.Context) (*alertlog.Entry, error) {
	return e.logDB.FindLatestByType(ctx, e.alertID, e._type)
}

func (e logError) Error() string {
	if e.isAlreadyAcknowledged {
		return "alert is already acknowledged"
	}
	if e.isAlreadyClosed {
		return "alert is already closed"
	}
	return "unknown status update"
}

func AlertID(err error) int {
	var e logError
	if errors.As(err, &e) {
		return e.alertID
	}

	return 0
}

func IsAlreadyAcknowledged(err error) bool {
	var e logError
	if errors.As(err, &e) {
		return e.isAlreadyAcknowledged
	}
	return false
}

func IsAlreadyClosed(err error) bool {
	var e logError
	if errors.As(err, &e) {
		return e.isAlreadyClosed
	}
	return false
}
