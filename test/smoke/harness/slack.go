package harness

import (
	"fmt"
	"net/http/httptest"
	"sort"
	"strings"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/devtools/mockslack"
)

const (
	SlackTestSigningSecret = "secret"
)

type SlackServer interface {
	Channel(string) SlackChannel
	User(string) SlackUser
	UserGroup(string) SlackUserGroup

	WaitAndAssert()
}

type SlackUser interface {
	ID() string
	Name() string

	ExpectMessage(keywords ...string) SlackMessage
}

type SlackChannel interface {
	ID() string
	Name() string

	ExpectMessage(keywords ...string) SlackMessage
	ExpectEphemeralMessage(keywords ...string) SlackMessage
}

type SlackUserGroup interface {
	ID() string
	Name() string
	ErrorChannel() SlackChannel

	ExpectUsers(names ...string)
	ExpectUserIDs(ids ...string)
}

type SlackMessageState interface {
	// AssertText asserts that the message contains the given keywords.
	AssertText(keywords ...string)

	// AssertNotText asserts that the message does not contain the given keywords.
	AssertNotText(keywords ...string)

	// AssertColor asserts that the message has the given color bar value.
	AssertColor(color string)

	// AssertActions asserts that the message includes the given action buttons.
	AssertActions(labels ...string)

	// Action returns the action with the given label.
	Action(label string) SlackAction
}

type SlackAction interface {
	Click()
	URL() string
}

type SlackMessage interface {
	SlackMessageState

	ExpectUpdate() SlackMessageState

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
	ug         map[string]*slackUserGroup
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

type slackAction struct {
	*slackMessage
	mockslack.Action
}

func (msg *slackMessage) AssertActions(text ...string) {
	msg.h.t.Helper()

	require.Equalf(msg.h.t, len(text), len(msg.Actions), "message actions")
	sort.Slice(text, func(i, j int) bool { return text[i] < text[j] })
	sort.Slice(msg.Actions, func(i, j int) bool { return msg.Actions[i].Text < msg.Actions[j].Text })
	for i, a := range msg.Actions {
		require.Equalf(msg.h.t, text[i], a.Text, "message action text")
	}
}

func (msg *slackMessage) Action(text string) SlackAction {
	msg.h.t.Helper()

	var a *mockslack.Action
	for _, action := range msg.Actions {
		if action.Text != text {
			continue
		}
		a = &action
		break
	}
	require.NotNilf(msg.h.t, a, `expected action "%s"; got %#v`, text, msg.Actions)
	msg.h.t.Logf("found action: %s\n%#v", text, *a)

	return &slackAction{
		slackMessage: msg,
		Action:       *a,
	}
}

func (a *slackAction) URL() string {
	a.h.t.Helper()
	return a.Action.URL
}

func (a *slackAction) Click() {
	a.h.t.Helper()

	a.h.t.Logf("clicking action: %s", a.Text)
	asID := a.h.slackUser.ID
	if strings.HasPrefix(a.ChannelID, "W") {
		// Perform actions in DMs as the user who received the DM.
		asID = a.ChannelID
	}
	err := a.h.slack.PerformActionAs(asID, a.Action)
	require.NoError(a.h.t, err, "perform Slack action")
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

func (s *slackServer) User(name string) SlackUser {
	ch := s.channels["_user:"+name]
	if ch != nil {
		return ch
	}

	info := s.NewUser(name)

	ch = &slackChannel{h: s.h, name: "@" + name, id: info.ID}
	s.channels["_user:"+name] = ch

	return ch
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

type slackUserGroup struct {
	h       *Harness
	name    string
	ugID    string
	channel SlackChannel
}

func (s *slackServer) UserGroup(name string) SlackUserGroup {
	ug := s.ug[name]
	if ug != nil {
		return ug
	}

	mUG := s.NewUserGroup(name)
	ch := s.Channel("ug:" + name)

	ug = &slackUserGroup{h: s.h, name: fmt.Sprintf("@%s (%s)", name, ch.Name()), ugID: mUG.ID, channel: ch}

	s.ug[name] = ug

	return ug
}

func (ug *slackUserGroup) ID() string                 { return ug.ugID + ":" + ug.channel.ID() }
func (ug *slackUserGroup) Name() string               { return ug.name }
func (ug *slackUserGroup) ErrorChannel() SlackChannel { return ug.channel }

func (ug *slackUserGroup) ExpectUsers(names ...string) {
	ug.h.t.Helper()

	var ids []string
	for _, name := range names {
		ids = append(ids, ug.h.Slack().User(name).ID())
	}
	ug.ExpectUserIDs(ids...)
}

func (ug *slackUserGroup) ExpectUserIDs(ids ...string) {
	ug.h.t.Helper()

	require.EventuallyWithT(ug.h.t, func(t *assert.CollectT) {
		if assert.ElementsMatch(t, ug.h.slack.UserGroupUserIDs(ug.ugID), ids, "List A = expected; List B = actual") {
			return
		}

		ug.h.Trigger()
	}, 15*time.Second, time.Millisecond, "UserGroup Users should match")
}

func (ch *slackChannel) ID() string   { return ch.id }
func (ch *slackChannel) Name() string { return ch.name }

func (ch *slackChannel) ExpectMessage(keywords ...string) SlackMessage {
	ch.h.t.Helper()
	return ch.expectMessageFunc("message", func(msg mockslack.Message) bool {
		// only return non-thread replies
		return msg.ThreadTS == ""
	}, keywords...)
}

func (ch *slackChannel) ExpectEphemeralMessage(keywords ...string) SlackMessage {
	ch.h.t.Helper()
	return ch.expectMessageFunc("ephemeral", func(msg mockslack.Message) bool {
		// only return non-thread replies
		return msg.ToUserID != ""
	}, keywords...)
}
func containsAllKeywords(text string, keywords ...string) bool {
	for _, w := range keywords {
		if !strings.Contains(text, w) {
			return false
		}
	}
	return true
}

func (ch *slackChannel) expectMessageFunc(desc string, test func(mockslack.Message) bool, keywords ...string) (found *slackMessage) {
	ch.h.t.Helper()

	ch.h.Trigger()
	require.Eventually(ch.h.t, func() bool {
		for _, msg := range ch.h.slack.Messages(ch.id) {
			if !test(msg) {
				continue
			}
			if !containsAllKeywords(msg.Text, keywords...) {
				continue
			}
			ch.h.t.Logf("received Slack message to %s: %s", ch.name, msg.Text)
			ch.h.slack.DeleteMessage(ch.id, msg.TS)

			found = &slackMessage{
				h:       ch.h,
				channel: ch,
				Message: msg,
			}
			return true
		}
		return false
	}, 30*time.Second, time.Second, "expected to find Slack %s: Channel=%s; ID=%s; keywords=%v", desc, ch.name, ch.id, keywords)
	return found
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

func (msg *slackMessage) AssertColor(color string) {
	msg.h.t.Helper()

	if msg.Color != color {
		require.Equalf(msg.h.t, color, msg.Color, "message color")
	}
}

func (msg *slackMessage) AssertText(keywords ...string) {
	msg.h.t.Helper()

	for _, w := range keywords {
		require.Contains(msg.h.t, msg.Text, w)
	}
}

func (msg *slackMessage) AssertNotText(keywords ...string) {
	msg.h.t.Helper()

	for _, w := range keywords {
		require.NotContains(msg.h.t, msg.Text, w)
	}
}

func (msg *slackMessage) ExpectUpdate() SlackMessageState {
	msg.h.t.Helper()

	return msg.channel.expectMessageFunc("message update", func(m mockslack.Message) bool {
		return m.UpdateTS == msg.TS
	})
}

func (msg *slackMessage) ExpectThreadReply(keywords ...string) {
	msg.h.t.Helper()

	reply := msg.channel.expectMessageFunc("thread reply", func(m mockslack.Message) bool {
		return m.ThreadTS == msg.TS
	}, keywords...)

	assert.False(msg.h.t, reply.Broadcast, "expected thread reply to not be broadcast")
}

func (msg *slackMessage) ExpectBroadcastReply(keywords ...string) {
	msg.h.t.Helper()

	reply := msg.channel.expectMessageFunc("broadcast reply", func(m mockslack.Message) bool {
		return m.ThreadTS == msg.TS
	}, keywords...)

	assert.True(msg.h.t, reply.Broadcast, "expected thread reply to be broadcast")
}

func (h *Harness) initSlack() {
	h.slack = &slackServer{
		h:        h,
		channels: make(map[string]*slackChannel),
		ug:       make(map[string]*slackUserGroup),
		Server:   mockslack.NewServer(),
	}
	h.slackS = httptest.NewServer(h.slack)

	app, err := h.slack.InstallStaticApp(mockslack.AppInfo{Name: "GoAlert Smoketest", SigningSecret: SlackTestSigningSecret}, "bot")
	require.NoError(h.t, err)
	h.slackApp = *app
	h.slackUser = h.slack.NewUser("GoAlert Smoketest User")

	h.slack.SetURLPrefix(h.slackS.URL)
}
