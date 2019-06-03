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
	userLim := newConcurrencyLimiter(cfg.perUser)
	svcLim := newConcurrencyLimiter(cfg.perService)
	intLim := newConcurrencyLimiter(cfg.perIntKey)
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		if id := getIntKey(ctx); id != "" {
			err := intLim.Lock(ctx, id)
			if err != nil {
				log.Debug(ctx, err)
				return
			}
			defer intLim.Unlock(id)
		}

		if id := permission.ServiceID(ctx); id != "" {
			err := svcLim.Lock(ctx, id)
			if err != nil {
				log.Debug(ctx, err)
				return
			}
			defer svcLim.Unlock(id)
		}

		if id := permission.UserID(ctx); id != "" {
			err := userLim.Lock(ctx, id)
			if err != nil {
				log.Debug(ctx, err)
				return
			}
			defer userLim.Unlock(id)
		}

		next.ServeHTTP(w, req)
	})
}
