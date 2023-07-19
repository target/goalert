package harness

import (
	"context"
	"net"

	"github.com/emersion/go-smtp"
	"github.com/target/goalert/smtpsrv"
)

func newSMTPServer() (*smtp.Server, net.Listener) {
	var s *smtp.Server

	cfg := smtpsrv.Config{
		ListenAddr:     "localhost:0",
		AllowedDomains: []string{"smoketest.example.com"},
	}

	l, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		panic(err)
	}

	s = smtpsrv.NewServer(&cfg)

	h := smtpsrv.IngressSMTP(nil, nil, &cfg)

	go func() {
		h.ServeSMTP(context.Background(), s, l)
	}()

	return s, l
}
