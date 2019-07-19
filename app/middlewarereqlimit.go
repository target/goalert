package app

import (
	"context"
	"net/http"
	"time"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
)

type conReqLimit struct {
	perUser    int
	perService int
	perIntKey  int
}

func getIntKey(ctx context.Context) string {
	src := permission.Source(ctx)
	if src != nil && src.Type == permission.SourceTypeIntegrationKey {
		return src.ID
	}
	return ""
}
func (cfg conReqLimit) Middleware(next http.Handler) http.Handler {
	userLim := newConcurrencyLimiter(cfg.perUser, 250)
	svcLim := newConcurrencyLimiter(cfg.perService, 250)
	intLim := newConcurrencyLimiter(cfg.perIntKey, 250)
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		lockCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
		defer cancel()

		failure := func(err error) bool {
			if err == nil {
				return false
			}

			log.Debug(ctx, err)
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return true
		}

		if id := getIntKey(ctx); id != "" {
			if failure(intLim.Lock(lockCtx, id)) {
				return
			}
			defer intLim.Unlock(id)
		}

		if id := permission.ServiceID(ctx); id != "" {
			if failure(svcLim.Lock(ctx, id)) {
				return
			}
			defer svcLim.Unlock(id)
		}

		if id := permission.UserID(ctx); id != "" {
			if failure(userLim.Lock(ctx, id)) {
				return
			}
			defer userLim.Unlock(id)
		}

		next.ServeHTTP(w, req)
	})
}
