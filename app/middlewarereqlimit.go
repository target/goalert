package app

import (
	"context"
	"net/http"

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
	usrLim := newConcurrencyLimiter(cfg.perUser)
	svcLim := newConcurrencyLimiter(cfg.perService)
	intLim := newConcurrencyLimiter(cfg.perIntKey)
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		var lim *concurrencyLimiter
		var id string
		if key := getIntKey(ctx); key != "" {
			err := intLim.Lock(ctx, key)
			if err != nil {
				log.Debug(ctx, err)
				return
			}
			defer intLim.Unlock(key)
		}
		if svc := permission.ServiceID(ctx); svc != "" {
			lim, id = svcLim, svc
		}
		if uid := permission.UserID(ctx); uid != "" {
			lim, id = usrLim, uid
		}

		if lim == nil {
			next.ServeHTTP(w, req)
			return
		}

		err := lim.Lock(ctx, id)
		if err != nil {
			// context canceled
			log.Debug(ctx, err)
			return
		}
		defer lim.Unlock(id)
		next.ServeHTTP(w, req)
	})
}
