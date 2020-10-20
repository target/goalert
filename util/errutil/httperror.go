package errutil

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/pkg/errors"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
)

func isCtxCause(err error) bool {
	if errors.Is(err, context.Canceled) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
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
	if permission.IsUnauthorized(err) {
		log.Debug(ctx, err)
		http.Error(w, unwrapAll(err).Error(), http.StatusUnauthorized)
		return true
	}
	if permission.IsPermissionError(err) {
		log.Debug(ctx, err)
		http.Error(w, unwrapAll(err).Error(), http.StatusForbidden)
		return true
	}
	if validation.IsClientError(err) {
		log.Debug(ctx, err)
		http.Error(w, unwrapAll(err).Error(), http.StatusBadRequest)
		return true
	}
	if IsLimitError(err) {
		log.Debug(ctx, err)
		http.Error(w, unwrapAll(err).Error(), http.StatusConflict)
		return true
	}

	if ctx.Err() != nil && isCtxCause(err) {
		// context timed out or was canceled
		log.Debug(ctx, err)
		http.Error(w, http.StatusText(http.StatusGatewayTimeout), http.StatusGatewayTimeout)
		return true
	}

	log.Log(ctx, err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	return true
}
