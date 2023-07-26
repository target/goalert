package smtpsrv

import (
	"context"
	"net"
	"time"

	"github.com/emersion/go-smtp"
)

type Server struct {
	cfg Config
	srv *smtp.Server
}

func NewServer(cfg Config) *Server {
	s := &Server{cfg: cfg}

	srv := smtp.NewServer(s)
	srv.Domain = cfg.Domain
	srv.ReadTimeout = 10 * time.Second
	srv.WriteTimeout = 10 * time.Second
	srv.AuthDisabled = true
	srv.TLSConfig = cfg.TLSConfig
	s.srv = srv

	return s
}

func (s *Server) ServeSMTP(l net.Listener) error { return s.srv.Serve(l) }

func (s *Server) NewSession(_ *smtp.Conn) (smtp.Session, error) { return &Session{cfg: s.cfg}, nil }

func (s *Server) Shutdown(ctx context.Context) error { return s.srv.Shutdown(ctx) }
