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
	if err == context.Canceled {
		return true
	}
	if err == context.DeadlineExceeded {
		return true
	}
	if err == sql.ErrTxDone {
		return true
	}

	// 57014 = query_canceled
	// https://www.postgresql.org/docs/9.6/static/errcodes-appendix.html
	if e := sqlutil.MapError(err); e != nil && e.Code == "57014" {
		return true
	}

	return false
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
		http.Error(w, errors.Cause(err).Error(), http.StatusUnauthorized)
		return true
	}
	if permission.IsPermissionError(err) {
		log.Debug(ctx, err)
		http.Error(w, errors.Cause(err).Error(), http.StatusForbidden)
		return true
	}
	if validation.IsClientError(err) {
		log.Debug(ctx, err)
		http.Error(w, errors.Cause(err).Error(), http.StatusBadRequest)
		return true
	}
	if IsLimitError(err) {
		log.Debug(ctx, err)
		http.Error(w, errors.Cause(err).Error(), http.StatusConflict)
		return true
	}

	if ctx.Err() != nil && isCtxCause(errors.Cause(err)) {
		// context timed out or was canceled
		log.Debug(ctx, err)
		http.Error(w, http.StatusText(http.StatusGatewayTimeout), http.StatusGatewayTimeout)
		return true
	}

	log.Log(ctx, err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	return true
}
