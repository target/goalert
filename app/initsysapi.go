package app

import (
	"context"
	"net"

	"github.com/target/goalert/sysapi"
	"github.com/target/goalert/sysapi/sysapiapp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
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
	reflection.Register(srv)
	sysapi.RegisterSysAPIServer(srv, &sysapiapp.Server{UserStore: app.UserStore})
	app.hSrv = health.NewServer()
	grpc_health_v1.RegisterHealthServer(srv, app.hSrv)

	app.sysAPISrv = srv
	app.sysAPIL = lis
	return nil
}
