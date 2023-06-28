package mockslack

import (
	"sync"
	"time"
)

type state struct {
	mx sync.Mutex

	gen *idGen

	flags struct {
		autoCreateChannel bool
	}

	apps       map[string]*appState
	channels   map[string]*channelState
	tokens     map[string]*AuthToken
	users      map[string]*userState
	tokenCodes map[string]*tokenCode
	usergroups map[string]*usergroupState
	teamID     string
}

func newState() *state {
	return &state{
		gen:        newIDGen(),
		apps:       make(map[string]*appState),
		channels:   make(map[string]*channelState),
		tokens:     make(map[string]*AuthToken),
		users:      make(map[string]*userState),
		tokenCodes: make(map[string]*tokenCode),
		usergroups: make(map[string]*usergroupState),
		teamID:     genTeamID(),
	}
}

type tokenCode struct {
	ClientID string
	*AuthToken
}

type appState struct {
	App
}
type App struct {
	ID        string
	Name      string
	Secret    string
	AuthToken *AuthToken
	ActionURL string

	SigningSecret string
}

type channelState struct {
	Channel

	TS       time.Time
	Users    []string
	Messages []*Message
}

type usergroupState struct {
	UserGroup

	Users []string
}

// SetAutoCreateChannel, if set to true, will cause messages sent to
// non-existent channels to succeed by creating the channel automatically.
func (st *state) SetAutoCreateChannel(value bool) {
	st.mx.Lock()
	defer st.mx.Unlock()

	st.flags.autoCreateChannel = value
}

func (st *state) token(id string) *AuthToken {
	st.mx.Lock()
	defer st.mx.Unlock()

	return st.tokens[id]
}
func (st *state) app(id string) *appState {
	st.mx.Lock()
	defer st.mx.Unlock()

	return st.apps[id]
}

func (st *state) newToken(a AuthToken) *AuthToken {
	st.mx.Lock()
	defer st.mx.Unlock()
	if a.ID == "" {
		if a.IsBot {
			a.ID = st.gen.BotAccessToken()
		} else {
			a.ID = st.gen.UserAccessToken()
		}
	}
	st.tokens[a.ID] = &a
	return &a
}
