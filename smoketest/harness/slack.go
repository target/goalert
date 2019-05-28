package harness

import (
	"net/http/httptest"
	"strings"
	"time"

	"github.com/target/goalert/devtools/mockslack"
)

type SlackServer interface {
	Channel(string) SlackChannel

	WaitAndAssert()
}

type SlackChannel interface {
	ID() string
	Name() string

	ExpectMessage(keywords ...string)
}

type slackServer struct {
	h *Harness
	*mockslack.Server
	channels map[string]*slackChannel
}

type slackChannel struct {
	h    *Harness
	name string
	id   string

	expected [][]string
}

func (h *Harness) Slack() SlackServer { return h.slack }

func (s *slackServer) WaitAndAssert() {
	timeout := time.NewTimer(15 * time.Second)
	defer timeout.Stop()

	t := time.NewTicker(time.Millisecond)
	defer t.Stop()

	for _, ch := range s.channels {
		for !ch.waitAndAssert(timeout.C) {
			<-t.C
		}
	}
}

func (s *slackServer) Channel(name string) SlackChannel {
	ch := s.channels[name]
	if ch != nil {
		return ch
	}

	info := s.NewChannel(name)

	ch = &slackChannel{h: s.h, name: "#" + name, id: info.ID}
	s.channels[name] = ch

	return ch
}

func (ch *slackChannel) ID() string   { return ch.id }
func (ch *slackChannel) Name() string { return ch.name }
func (ch *slackChannel) ExpectMessage(keywords ...string) {
	ch.expected = append(ch.expected, keywords)
}

func (ch *slackChannel) waitAndAssert(timeout <-chan time.Time) bool {
	msgs := ch.h.slack.Messages(ch.id)

	check := func(keywords []string) bool {
	msgLoop:
		for i, msg := range msgs {
			for _, w := range keywords {
				if !strings.Contains(msg.Text, w) {
					continue msgLoop
				}
			}
			msgs = append(msgs[:i], msgs[i+1:]...)
			return true
		}
		return false
	}

	for i, exp := range ch.expected {
		select {
		case <-timeout:
			ch.h.t.Fatalf("timeout waiting for slack message: channel=%s; ID=%s; message=%d keywords=%v\nGot: %s", ch.name, ch.id, i, exp, msgs)
		default:
		}
		if !check(exp) {
			return false
		}
	}

	return true
}

func (h *Harness) initSlack() {

	h.slack = &slackServer{
		h:        h,
		channels: make(map[string]*slackChannel),
		Server:   mockslack.NewServer(),
	}
	h.slackS = httptest.NewServer(h.slack)

	h.slackApp = h.slack.InstallApp("GoAlert Smoketest", "bot")
	h.slackUser = h.slack.NewUser("GoAlert Smoketest User")
}
