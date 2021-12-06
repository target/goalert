package app

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/pkg/errors"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
)

type _reqInfoCtxKey string

const reqInfoCtxKey = _reqInfoCtxKey("request-info-fields")

func maxBodySizeMiddleware(size int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if size == 0 {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, size)
			next.ServeHTTP(w, r)
		})
	}
}

type readLogger struct {
	io.ReadCloser
	n int
}

func (r *readLogger) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	r.n += n
	return n, err
}

func logRequestAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		extraFields := req.Context().Value(reqInfoCtxKey).(*log.Fields)
		*extraFields = log.ContextFields(req.Context())
		next.ServeHTTP(w, req)
	})
}

func logRequest(alwaysLog bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			ctx = log.SetRequestID(ctx)
			ctx = log.WithFields(ctx, log.Fields{
				"http_method":      req.Method,
				"http_proto":       req.Proto,
				"remote_addr":      req.RemoteAddr,
				"host":             req.Host,
				"uri":              req.URL.Path,
				"referer":          req.Referer(),
				"x_forwarded_for":  req.Header.Get("x-forwarded-for"),
				"x_forwarded_host": req.Header.Get("x-forwarded-host"),
			})

			// Logging auth info in request
			ctx = context.WithValue(ctx, reqInfoCtxKey, &log.Fields{})

			rLog := &readLogger{ReadCloser: req.Body}
			req.Body = rLog

			var serveError interface{}
			metrics := httpsnoop.CaptureMetricsFn(w, func(w http.ResponseWriter) {
				defer func() {
					serveError = recover()
				}()
				next.ServeHTTP(w, req.WithContext(ctx))
			})

			if serveError != nil && metrics.Written == 0 {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				metrics.Code = 500
			}

			checks, _ := permission.AuthCheckCount(ctx)

			extraFields := ctx.Value(reqInfoCtxKey).(*log.Fields)
			ctx = log.WithFields(ctx, *extraFields)
			status := metrics.Code
			if status == 0 {
				status = 200
			}
			ctx = log.WithFields(ctx, log.Fields{
				"resp_bytes_length": metrics.Written,
				"req_bytes_length":  rLog.n,
				"resp_elapsed_ms":   metrics.Duration.Seconds() * 1000,
				"resp_status":       status,
				"AuthCheckCount":    checks,
			})

			if serveError != nil {
				switch e := serveError.(type) {
				case error:
					log.Log(ctx, errors.Wrap(e, "request panic"))
				default:
					log.Log(ctx, errors.Errorf("request panic: %v", e))
				}
				return
			}
			if alwaysLog && req.URL.Path != "/health" {
				log.Logf(ctx, "request complete")
			} else {
				log.Debugf(ctx, "request complete")
			}
		})
	}
}

func authCheckLimit(max int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			next.ServeHTTP(w, req.WithContext(
				permission.AuthCheckCountContext(req.Context(), uint64(max)),
			))
		})
	}
}

func timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, cancel := context.WithTimeout(req.Context(), timeout)
			defer cancel()
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}
