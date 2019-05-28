package alert

import (
	"context"
	alertlog "github.com/target/goalert/alert/log"

	"github.com/pkg/errors"
)

type LogEntryFetcher interface {
	// LogEntry fetchs the latest log entry for a given alertID and type.
	LogEntry(ctx context.Context) (alertlog.Entry, error)
}

type logError struct {
	isAlreadyAcknowledged bool
	isAlreadyClosed       bool
	alertID               int
	_type                 alertlog.Type
	logDB                 alertlog.Store
}

func (e logError) LogEntry(ctx context.Context) (alertlog.Entry, error) {
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

func IsAlreadyAcknowledged(err error) bool {
	if e, ok := errors.Cause(err).(logError); ok {
		return e.isAlreadyAcknowledged
	}
	return false
}

func IsAlreadyClosed(err error) bool {
	if e, ok := errors.Cause(err).(logError); ok {
		return e.isAlreadyClosed
	}
	return false
}
