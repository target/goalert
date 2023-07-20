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
		AllowedDomains: []string{"smoketest.example.com"},
	}

	l, err := net.Listen("tcp", "localhost:0")
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
