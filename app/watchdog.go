package app

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/app/lifecycle"
	"github.com/target/goalert/util/log"
)

type upgrader interface {
	Stop()
	Ready() error
	Exit() <-chan struct{}
	Upgrade() error
}

func (app *App) listenNoUpgrade() error {
	var err error
	app.l, err = net.Listen("tcp", app.cfg.ListenAddr)
	return err
}

func (app *App) watchdog(ctx context.Context) {
	if app.upg == nil {
		return
	}

	log.Logf(ctx, "Engine watchdog started.")
	var u url.URL
	u.Path = "/health/engine"
	u.Host = app.l.Addr().String()
	u.Scheme = "http"

	t := time.NewTicker(15 * time.Second)
	defer t.Stop()
	for range t.C {
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			panic(err)
		}
		ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		req = req.WithContext(ctx)
		resp, err := http.DefaultClient.Do(req)
		cancel()
		if err == nil && resp.StatusCode != 200 {
			err = errors.New("non-200 from engine health: " + resp.Status)
		}
		if err != nil {
			log.Log(ctx, errors.Wrap(err, "engine watchdog"))
			if app.Status() == lifecycle.StatusReady {
				log.Logf(ctx, "Failed engine check, restarting...")
				err = app.upg.Upgrade()
				if err != nil {
					log.Log(ctx, errors.Wrap(err, "watchdog upgrade"))
					continue
				}

				<-app.upg.Exit()
				log.Log(ctx, app.Shutdown(ctx))
				return
			}
		}
	}

}
