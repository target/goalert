package harness

import (
	"net"
	"strings"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type EmailServer interface {
	ExpectMessage(address string, keywords ...string)

	WaitAndAssert()
}

type emailServer struct {
	h *Harness

	mp *mailpit
}

func findOpenPorts(num int) ([]string, error) {
	var listeners []net.Listener
	for range num {
		ln, err := net.Listen("tcp", "localhost:0")
		if err != nil {
			for _, l := range listeners {
				l.Close()
			}
			return nil, err
		}
		listeners = append(listeners, ln)
	}

	var addrs []string
	for _, l := range listeners {
		addrs = append(addrs, l.Addr().String())
		l.Close()
	}

	return addrs, nil
}

func isListening(addr string) bool {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return false
	}
	c.Close()
	return true
}

func newEmailServer(h *Harness) *emailServer {
	mp, err := newMailpit(5)
	require.NoError(h.t, err)

	h.t.Logf("mailpit: smtp: %s", mp.smtpAddr)
	h.t.Logf("mailpit: api: %s", mp.apiAddr)

	return &emailServer{
		h:  h,
		mp: mp,
	}
}
func (e *emailServer) Close() error { return e.mp.Close() }
func (e *emailServer) Addr() string { return e.mp.smtpAddr }

func (h *Harness) Email(id string) string { return h.emailG.Get(id) }

func (h *Harness) SMTP() EmailServer { return h.email }

func (e *emailServer) ExpectMessage(address string, keywords ...string) {
	e.h.t.Helper()

	gotMessage := assert.Eventuallyf(e.h.t, func() bool {
		found, err := e.mp.ReadMessage(address, keywords...)
		require.NoError(e.h.t, err)
		return found
	}, 15*time.Second, 10*time.Millisecond, "expected to find email: address=%s; keywords=%v", address, keywords)
	if gotMessage {
		return
	}

	msgs, err := e.mp.UnreadMessages()
	assert.NoError(e.h.t, err)
	e.h.t.Fatalf("timeout waiting for email; Got:\n%v", msgs)
}

type emailMessage struct {
	address []string
	body    string
}

func (e *emailServer) WaitAndAssert() {
	e.h.t.Helper()

	msgs, err := e.mp.UnreadMessages()
	require.NoError(e.h.t, err)

	for _, msg := range msgs {
		e.h.t.Errorf("unexpected message: to=%s; body=%s", strings.Join(msg.address, ","), msg.body)
	}
}
