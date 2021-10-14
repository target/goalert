package harness

import (
	"net/http/httptest"
	"strings"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/devtools/mockslack"
)

type SlackServer interface {
	Channel(string) SlackChannel

	WaitAndAssert()
}

type SlackChannel interface {
	ID() string
	Name() string

	ExpectMessage(keywords ...string) SlackMessage
}

type SlackMessageState interface {
	// AssertText asserts that the message contains the given keywords.
	AssertText(keywords ...string)
}

type SlackMessage interface {
	SlackMessageState

	// ExpectThreadReply waits and asserts that a non-broadcast thread reply is received.
	ExpectThreadReply(keywords ...string)

	// ExpectBroadcastReply waits and asserts that a broadcast thread reply is received.
	ExpectBroadcastReply(keywords ...string)
}

type slackServer struct {
	h *Harness
	*mockslack.Server
	hasFailure bool
	channels   map[string]*slackChannel
}

type slackChannel struct {
	h    *Harness
	name string
	id   string
}

type slackMessage struct {
	h *Harness

	channel *slackChannel

	mockslack.Message
}

func (h *Harness) Slack() SlackServer { return h.slack }

func (s *slackServer) WaitAndAssert() {
	s.h.t.Helper()
	timeout := time.NewTimer(15 * time.Second)
	defer timeout.Stop()

	t := time.NewTicker(time.Millisecond)
	defer t.Stop()

	for i := 0; i < 3; i++ {
		s.h.Trigger()
		var hasFailure bool
		for _, ch := range s.channels {
			hasFailure = s.hasFailure || hasFailure || ch.hasUnexpectedMessages()
		}

		if hasFailure {
			s.h.t.FailNow()
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

func (ch *slackChannel) ExpectMessage(keywords ...string) SlackMessage {
	ch.h.t.Helper()
	return ch.expectMessageFunc(func(msg mockslack.Message) bool {
		// only return non-thread replies
		return msg.ThreadTS == ""
	}, keywords...)
}

func (ch *slackChannel) expectMessageFunc(test func(mockslack.Message) bool, keywords ...string) *slackMessage {
	ch.h.t.Helper()

	timeout := time.NewTimer(15 * time.Second)
	defer timeout.Stop()

	for {
	msgLoop:
		for _, msg := range ch.h.slack.Messages(ch.id) {
			for _, w := range keywords {
				if !strings.Contains(msg.Text, w) {
					continue msgLoop
				}
			}
			if !test(msg) {
				continue msgLoop
			}

			ch.h.t.Logf("received Slack message to %s: %s", ch.name, msg.Text)
			ch.h.slack.DeleteMessage(ch.id, msg.TS)

			return &slackMessage{
				h:       ch.h,
				channel: ch,
				Message: msg,
			}
		}

		select {
		case <-timeout.C:
			ch.h.slack.hasFailure = true
			ch.h.t.Fatalf("timeout waiting for slack message: Channel=%s; ID=%s; keywords=%v\nGot: %#v", ch.name, ch.id, keywords, ch.h.slack.Messages(ch.id))
			return nil
		default:
		}

		ch.h.Trigger()
	}
}

func (ch *slackChannel) hasUnexpectedMessages() bool {
	ch.h.t.Helper()

	var hasFailure bool
	for _, msg := range ch.h.slack.Messages(ch.id) {
		ch.h.t.Errorf("unexpected slack message: Channel=%s; ID=%s; Text=%s", ch.name, ch.id, msg.Text)
		hasFailure = true
	}

	return hasFailure
}

func (msg *slackMessage) AssertText(keywords ...string) {
	msg.h.t.Helper()

	for _, w := range keywords {
		if !strings.Contains(msg.Text, w) {
			msg.h.t.Errorf("slack message '%s' does not contain keyword: %s", msg.Text, w)
		}
	}
}
func (msg *slackMessage) ExpectThreadReply(keywords ...string) {
	msg.h.t.Helper()

	reply := msg.channel.expectMessageFunc(func(m mockslack.Message) bool {
		return m.ThreadTS == msg.TS
	}, keywords...)

	assert.False(msg.h.t, reply.Broadcast, "expected thread reply to not be broadcast")
}

func (msg *slackMessage) ExpectBroadcastReply(keywords ...string) {
	msg.h.t.Helper()

	reply := msg.channel.expectMessageFunc(func(m mockslack.Message) bool {
		return m.ThreadTS == msg.TS
	}, keywords...)

	assert.True(msg.h.t, reply.Broadcast, "expected thread reply to be broadcast")
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
