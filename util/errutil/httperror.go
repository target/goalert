package errutil

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/target/goalert/ctxlock"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
)

func isCancel(err error) bool {
	if errors.Is(err, context.Canceled) {
		return true
	}
	if errors.Is(err, sql.ErrTxDone) {
		return true
	}

	// 57014 = query_canceled
	// https://www.postgresql.org/docs/9.6/static/errcodes-appendix.html
	if e := sqlutil.MapError(err); e != nil && e.Code == "57014" {
		return true
	}

	return false
}

func unwrapAll(err error) error {
	for {
		next := errors.Unwrap(err)
		if next == nil {
			break
		}
		err = next
	}
	return err
}

// HTTPError will respond in a standard way when err != nil. If
// err is nil, false is returned, true otherwise.
func HTTPError(ctx context.Context, w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}

	err = MapDBError(err)
	switch {
	case errors.Is(err, ctxlock.ErrQueueFull), errors.Is(err, ctxlock.ErrTimeout):
		// Either the queue is full or the lock timed out. Either way
		// we are waiting on concurrent requests for this source, so
		// send them back with a 429 because we are rate limiting them
		// due to being at/beyond capacity.
		//
		// Because of the way the lock works, we can guarantee that
		// we will process one request at a time (per source), but we
		// may have to wait for a previous request to finish before we
		// can start processing the next one.
		//
		// This means only concurrent requests (per process) have the
		// possibility to be rate limited, and not sequential requests,
		// even in the worst case scenario.
		http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
	case isCancel(err):
		// Client disconnected, send 400 back so logs reflect that this
		// was a client-side problem.
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	case permission.IsUnauthorized(err):
		http.Error(w, unwrapAll(err).Error(), http.StatusUnauthorized)
	case permission.IsPermissionError(err):
		http.Error(w, unwrapAll(err).Error(), http.StatusForbidden)
	case validation.IsClientError(err):
		http.Error(w, unwrapAll(err).Error(), http.StatusBadRequest)
	case IsLimitError(err):
		http.Error(w, unwrapAll(err).Error(), http.StatusConflict)
	case errors.Is(err, context.DeadlineExceeded):
		// Timeout
		http.Error(w, http.StatusText(http.StatusGatewayTimeout), http.StatusGatewayTimeout)
	default:
		// For all other unexpected errors, log the error and send a 500.
		log.Log(ctx, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	return true
}
