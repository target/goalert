package smtpsrv

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/target/goalert/util/log"
)

// SMTPLogger implements the smtp.Logger interface using the main app Logger
type SMTPLogger struct {
	logger *log.Logger
}

func (l *SMTPLogger) Printf(format string, v ...interface{}) {
	l.logger.Error(context.Background(), errors.New(fmt.Sprintf(format, v...)))
}

func (l *SMTPLogger) Println(v ...interface{}) {
	l.logger.Error(context.Background(), errors.New(fmt.Sprint(v...)))
}

// Server implements an SMTP server that creates alerts.
type Server struct {
	cfg Config
	srv *smtp.Server
}

// NewServer creates a new Server.
func NewServer(cfg Config) *Server {
	s := &Server{cfg: cfg}

	srv := smtp.NewServer(s)
	srv.ErrorLog = &SMTPLogger{logger: cfg.Logger}
	srv.Domain = cfg.Domain
	srv.ReadTimeout = 10 * time.Second
	srv.WriteTimeout = 10 * time.Second
	srv.AuthDisabled = true
	srv.TLSConfig = cfg.TLSConfig
	s.srv = srv

	if cfg.MaxRecipients == 0 {
		cfg.MaxRecipients = 1
	}
	if cfg.BackgroundContext == nil {
		panic("smtpsrv: BackgroundContext is required")
	}
	if cfg.AuthorizeFunc == nil {
		panic("smtpsrv: AuthorizeFunc is required")
	}
	if cfg.CreateAlertFunc == nil {
		panic("smtpsrv: CreateAlertFunc is required")
	}

	return s
}

var _ smtp.Backend = &Server{}

// ServeSMTP starts the SMTP server on the given listener.
func (s *Server) ServeSMTP(l net.Listener) error { return s.srv.Serve(l) }

// NewSession implements the smtp.Backend interface.
func (s *Server) NewSession(_ *smtp.Conn) (smtp.Session, error) { return &Session{cfg: s.cfg}, nil }

// Shutdown gracefully shuts down the SMTP server.
func (s *Server) Shutdown(ctx context.Context) error { return s.srv.Shutdown(ctx) }
