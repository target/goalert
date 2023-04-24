package app

import (
	"net/http"
	"time"

	"github.com/target/goalert/ctxlock"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/errutil"
)

// LimitConcurrencyByAuthSource limits the number of concurrent requests
// per auth source. MaxHeld is 1, so only one request can be processed at a
// time per source (e.g., session key, integration key, etc).
//
// Note: This is per source/ID combo, so only multiple requests via the SAME
// integration key would get queued. Separate keys go in separate buckets.
func LimitConcurrencyByAuthSource(next http.Handler) http.Handler {
	limit := ctxlock.NewIDLocker[permission.SourceInfo](ctxlock.Config{
		MaxHeld: 1,
		MaxWait: 100,
		Timeout: 20 * time.Second,
	})

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		src := permission.Source(ctx)
		if src == nil {
			// Any unknown source gets put into a single bucket.
			src = &permission.SourceInfo{}
		}

		err := limit.Lock(ctx, *src)
		if errutil.HTTPError(ctx, w, err) {
			return
		}
		defer limit.Unlock(*src)

		next.ServeHTTP(w, req)
	})
}
