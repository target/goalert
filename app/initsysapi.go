package app

import (
	"context"
	"net"

	"github.com/target/goalert/sysapi"
	"github.com/target/goalert/sysapi/sysapiapp"
	"google.golang.org/grpc"
)

func (app *App) initSysAPI(ctx context.Context) error {
	if app.cfg.SysAPIListenAddr == "" {
		return nil
	}

	lis, err := net.Listen("tcp", app.cfg.SysAPIListenAddr)
	if err != nil {
		return err
	}

	srv := grpc.NewServer()
	sysapi.RegisterSysAPIServer(srv, &sysapiapp.Server{})

	app.sysAPISrv = srv
	app.sysAPIL = lis

	return nil
}
