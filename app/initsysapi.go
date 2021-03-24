package app

import (
	"context"
	"net"

	"github.com/target/goalert/sysapi"
	"github.com/target/goalert/sysapi/sysapiapp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func (app *App) initSysAPI(ctx context.Context) error {
	if app.cfg.SysAPIListenAddr == "" {
		return nil
	}

	lis, err := net.Listen("tcp", app.cfg.SysAPIListenAddr)
	if err != nil {
		return err
	}

	var opts []grpc.ServerOption
	if app.cfg.SysAPICertFile+app.cfg.SysAPIKeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(app.cfg.SysAPICertFile, app.cfg.SysAPIKeyFile)
		if err != nil {
			return err
		}
		opts = append(opts, grpc.Creds(creds))
	}

	srv := grpc.NewServer(opts...)
	sysapi.RegisterSysAPIServer(srv, &sysapiapp.Server{UserStore: app.UserStore})

	app.sysAPISrv = srv
	app.sysAPIL = lis

	return nil
}
