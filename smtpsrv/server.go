package smtpsrv

import (
	"context"
	"net"
	"time"

	"github.com/emersion/go-smtp"
)

// Server implements an SMTP server that creates alerts.
type Server struct {
	cfg Config
	srv *smtp.Server
}

// NewServer creates a new Server.
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

var _ smtp.Backend = &Server{}

// ServeSMTP starts the SMTP server on the given listener.
func (s *Server) ServeSMTP(l net.Listener) error { return s.srv.Serve(l) }

// NewSession implements the smtp.Backend interface.
func (s *Server) NewSession(_ *smtp.Conn) (smtp.Session, error) { return &Session{cfg: s.cfg}, nil }

// Shutdown gracefully shuts down the SMTP server.
func (s *Server) Shutdown(ctx context.Context) error { return s.srv.Shutdown(ctx) }
