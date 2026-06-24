package smtpsrv

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/emersion/go-smtp"
)

// SMTPLogger implements the smtp.Logger interface using the main app Logger.
type SMTPLogger struct {
	logger *slog.Logger
}

// Printf adheres to smtp.Server's Logger interface.
func (l *SMTPLogger) Printf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	// TODO: Uses string compare to filter out errors caused by TCP health checks,
	// remove once https://github.com/emersion/go-smtp/issues/236 has been fixed.
	if strings.Contains(s, "read: connection reset by peer") {
		return
	}
	l.logger.Error("SMTP error.", slog.String("error", s))
}

// Print adheres to smtp.Server's Logger interface.
func (l *SMTPLogger) Println(v ...interface{}) {
	s := fmt.Sprint(v...)
	// TODO: Uses string compare to filter out errors caused by TCP health checks,
	// remove once https://github.com/emersion/go-smtp/issues/236 has been fixed.
	if strings.Contains(s, "read: connection reset by peer") {
		return
	}
	l.logger.Error("SMTP error.", slog.String("error", s))
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
