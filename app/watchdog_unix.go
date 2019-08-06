// +build !windows

package app

import (
	"github.com/cloudflare/tableflip"
)

func (app *App) listen() error {
	if app.cfg.DisableWatchdog || app.cfg.APIOnly {
		return app.listenNoUpgrade()
	}
	upg, err := tableflip.New(tableflip.Options{})
	if err != nil {
		return err
	}
	app.upg = upg

	app.l, err = upg.Listen("tcp", app.cfg.ListenAddr)
	return err
}
