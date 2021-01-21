package harness

import (
	"net"
	"strings"
	"time"

	"github.com/mailhog/MailHog-Server/smtp"
	"github.com/mailhog/data"
	"github.com/mailhog/storage"
)

type EmailServer interface {
	ExpectMessage(address string, keywords ...string)

	WaitAndAssert()
}

type emailServer struct {
	h *Harness

	store *storage.InMemory
	l     net.Listener

	expected []emailExpect
}

type emailExpect struct {
	address  string
	keywords []string
}

func newEmailServer(h *Harness) *emailServer {
	store := storage.CreateInMemory()
	msgChan := make(chan *data.Message)
	go func() {
		// drain channel, TODO: check for leak
		for range msgChan {
		}
	}()

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}

			go smtp.Accept(
				conn.(*net.TCPConn).RemoteAddr().String(),
				conn,
				store,
				msgChan,
				"goalet-test.local",
				nil,
			)
		}
	}()

	return &emailServer{
		h:     h,
		store: store,
		l:     ln,
	}
}
func (e *emailServer) Close() error { return e.l.Close() }
func (e *emailServer) Addr() string { return e.l.Addr().String() }

func (h *Harness) Email(id string) string { return h.emailG.Get(id) }

func (h *Harness) SMTP() EmailServer { return h.email }

func (e *emailServer) ExpectMessage(address string, keywords ...string) {
	e.expected = append(e.expected, emailExpect{address: address, keywords: keywords})
}

type emailMessage struct {
	address []string
	body    string
}

func containsStr(s []string, search string) bool {
	for _, str := range s {
		if strings.Contains(str, search) {
			return true
		}
	}
	return false
}
func (e *emailServer) messages() []emailMessage {
	_msgs, err := e.store.List(0, 1000)
	if err != nil {
		panic(err)
	}
	msgs := []data.Message(*_msgs)

	var result []emailMessage
	for _, msg := range msgs {
		var addrs []string
		for _, p := range msg.To {
			addrs = append(addrs, p.Mailbox+"@"+p.Domain)
		}

		for _, part := range msg.MIME.Parts {
			if !containsStr(part.Headers["Content-Type"], "text/plain") {
				continue
			}
			result = append(result, emailMessage{
				body:    part.Body,
				address: addrs,
			})
		}
	}

	return result
}

func (e *emailServer) waitAndAssert(timeout <-chan time.Time) bool {
	msgs := e.messages()

	check := func(address string, keywords []string) bool {

	msgLoop:
		for i, msg := range msgs {
			var destMatch bool
			for _, addr := range msg.address {
				if addr == address {
					destMatch = true
					break
				}
			}
			if !destMatch {
				break
			}
			for _, w := range keywords {
				if !strings.Contains(msg.body, w) {
					continue msgLoop
				}
			}
			msgs = append(msgs[:i], msgs[i+1:]...)
			return true
		}
		return false
	}

	for i, exp := range e.expected {
		select {
		case <-timeout:
			e.h.t.Fatalf("timeout waiting for email: address=%s; message=%d keywords=%v\nGot: %s", exp.address, i, exp.keywords, msgs)
		default:
		}
		if !check(exp.address, exp.keywords) {
			return false
		}
	}

	for _, msg := range msgs {
		e.h.t.Errorf("unexpected message: to=%s; body=%s", strings.Join(msg.address, ","), msg.body)
	}

	return true
}

func (e *emailServer) WaitAndAssert() {
	timeout := time.NewTimer(15 * time.Second)
	defer timeout.Stop()

	t := time.NewTicker(time.Millisecond)
	defer t.Stop()

	for !e.waitAndAssert(timeout.C) {
		<-t.C
	}

	e.expected = nil
}
