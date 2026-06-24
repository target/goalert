package mockslack

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
)

// Server implements a mock Slack API.
type Server struct {
	*state

	mux *http.ServeMux

	handler http.Handler

	urlPrefix string
}

// NewServer creates a new blank Server.
func NewServer() *Server {
	srv := &Server{
		mux:   http.NewServeMux(),
		state: newState(),
	}

	srv.mux.HandleFunc("/actions/response", srv.ServeActionResponse)
	srv.mux.HandleFunc("/api/chat.postMessage", srv.ServeChatPostMessage)
	srv.mux.HandleFunc("/api/chat.postEphemeral", srv.ServeChatPostMessage)
	srv.mux.HandleFunc("/api/chat.update", srv.ServeChatUpdate)
	srv.mux.HandleFunc("/api/conversations.info", srv.ServeConversationsInfo)
	srv.mux.HandleFunc("/api/conversations.list", srv.ServeConversationsList)
	srv.mux.HandleFunc("/api/users.conversations", srv.ServeConversationsList) // same data
	srv.mux.HandleFunc("/api/users.info", srv.ServeUsersInfo)
	srv.mux.HandleFunc("/api/oauth.access", srv.ServeOAuthAccess)
	srv.mux.HandleFunc("/api/auth.revoke", srv.ServeAuthRevoke)
	srv.mux.HandleFunc("/api/auth.test", srv.ServeAuthTest)
	srv.mux.HandleFunc("/api/channels.create", srv.ServeChannelsCreate)
	srv.mux.HandleFunc("/api/groups.create", srv.ServeGroupsCreate)
	srv.mux.HandleFunc("/api/team.info", srv.ServeTeamInfo)
	srv.mux.HandleFunc("/api/usergroups.list", srv.ServeUserGroupList)
	srv.mux.HandleFunc("/api/usergroups.users.update", srv.ServeUserGroupsUsersUpdate)
	// TODO: history, leave, join

	srv.mux.HandleFunc("/stats", func(w http.ResponseWriter, req *http.Request) {
		srv.mx.Lock()
		defer srv.mx.Unlock()
		spew.Fdump(w)
	})

	// handle 404/unknown api methods
	srv.mux.HandleFunc("/api/", func(w http.ResponseWriter, req *http.Request) {
		err := json.NewEncoder(w).Encode(response{Err: "unknown_method: " + strings.TrimPrefix(req.URL.Path, "/api/")})
		if err != nil {
			log.Println("ERROR:", err)
		}
	})

	srv.mux.HandleFunc("/state", func(w http.ResponseWriter, req *http.Request) {
		srv.mx.Lock()
		defer srv.mx.Unlock()
		spew.Fdump(w, srv.state)
	})

	srv.handler = middleware(srv.mux,
		srv.tokenMiddleware,
		srv.loginMiddleware,
	)

	return srv
}

// SetURLPrefix will update the URL prefix for this server.
func (s *Server) SetURLPrefix(prefix string) {
	s.urlPrefix = prefix
}

// TokenCookieName is the name of a cookie containing a token for a user session.
const TokenCookieName = "slack_token"

// AppInfo contains information for an installed Slack app.
type AppInfo struct {
	Name         string
	ClientID     string
	ClientSecret string
	AccessToken  string
	TeamID       string

	SigningSecret string

	ActionURL string
}

func (s *Server) SetActionURL(appID string, actionURL string) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.apps[appID].ActionURL = actionURL
}

// InstallApp will "install" a new app to this Slack server using pre-configured AppInfo.
func (st *state) InstallStaticApp(app AppInfo, scopes ...string) (*AppInfo, error) {
	st.mx.Lock()
	defer st.mx.Unlock()

	if app.ClientID == "" {
		app.ClientID = st.gen.ClientID()
	}
	if app.ClientSecret == "" {
		app.ClientSecret = st.gen.ClientSecret()
	}
	if app.AccessToken == "" {
		app.AccessToken = st.gen.UserAccessToken()
	}
	app.TeamID = st.teamID

	if !clientIDRx.MatchString(app.ClientID) {
		return nil, errors.Errorf("invalid client ID format: %s", app.ClientID)
	}
	if !clientSecretRx.MatchString(app.ClientSecret) {
		return nil, errors.Errorf("invalid client secret format: %s", app.ClientSecret)
	}
	if !userAccessTokenRx.MatchString(app.AccessToken) {
		return nil, errors.Errorf("invalid access token format: %s", app.AccessToken)
	}

	for _, scope := range scopes {
		if !scopeRx.MatchString(scope) {
			panic("invalid scope format: " + scope)
		}
	}

	tok := &AuthToken{
		ID:     app.AccessToken,
		Scopes: scopes,
		User:   app.ClientID,
	}

	st.tokens[tok.ID] = tok
	st.apps[tok.User] = &appState{
		App: App{
			ID:        app.ClientID,
			Name:      app.Name,
			Secret:    app.ClientSecret,
			AuthToken: tok,
			ActionURL: app.ActionURL,

			SigningSecret: app.SigningSecret,
		},
	}

	return &app, nil
}

// InstallApp will "install" a new app to this Slack server.
func (st *state) InstallApp(name string, scopes ...string) AppInfo {
	app, err := st.InstallStaticApp(AppInfo{Name: name}, scopes...)
	if err != nil {
		// should not happen, since empty values are generated
		panic(err)
	}
	return *app
}

// UserInfo contains information for a newly created user.
type UserInfo struct {
	ID        string
	Name      string
	AuthToken string
}

// NewUser will create a new Slack user with the given name.
func (st *state) NewUser(name string) UserInfo {
	usr := st.newUser(User{Name: name})
	tok := st.newToken(AuthToken{
		User:   usr.ID,
		Scopes: []string{"user"},
	})

	return UserInfo{
		ID:        usr.ID,
		Name:      usr.Name,
		AuthToken: tok.ID,
	}
}

// ChannelInfo contains information about a newly created Slack channel.
type ChannelInfo struct {
	ID, Name string
}

// NewChannel will create a new Slack channel with the given name.
func (st *state) NewChannel(name string) ChannelInfo {
	info := ChannelInfo{
		ID:   st.gen.ChannelID(),
		Name: name,
	}

	st.mx.Lock()
	st.channels[info.ID] = &channelState{Channel: Channel{
		ID:        info.ID,
		Name:      info.Name,
		IsChannel: true,
	}}
	st.mx.Unlock()

	return info
}

// UserGroupInfo contains information about a newly created Slack user group.
type UserGroupInfo struct {
	ID, Name, Handle string
}

// NewUserGroup will create a new Slack user group with the given name.
func (st *state) NewUserGroup(name string) UserGroupInfo {
	info := UserGroupInfo{
		ID:     st.gen.UserGroupID(),
		Name:   name,
		Handle: name,
	}

	st.mx.Lock()
	st.usergroups[info.ID] = &usergroupState{UserGroup: UserGroup{
		ID:          info.ID,
		Name:        info.Name,
		Handle:      info.Handle,
		IsUserGroup: true,
	}}
	st.mx.Unlock()

	return info
}

// UserGroupUserIDs will return all users from a given user group.
func (st *state) UserGroupUserIDs(ugID string) []string {
	st.mx.Lock()
	defer st.mx.Unlock()

	ug := st.usergroups[ugID]
	if ug == nil {
		return nil
	}

	users := make([]string, len(ug.Users))
	copy(users, ug.Users)

	return users
}

// Messages will return all messages from a given channel/group.
func (st *state) Messages(chanID string) []Message {
	st.mx.Lock()
	defer st.mx.Unlock()
	ch := st.channels[chanID]
	if ch == nil {
		return nil
	}

	result := make([]Message, len(ch.Messages))
	for i, msg := range ch.Messages {
		result[i] = *msg
	}

	return result
}

// DeleteMessage will delete a message from channel history.
func (st *state) DeleteMessage(chanID, ts string) bool {
	st.mx.Lock()
	defer st.mx.Unlock()
	ch := st.channels[chanID]
	if ch == nil {
		return false
	}

	var deleted bool
	msgs := ch.Messages[:0]
	for _, m := range ch.Messages {
		if m.TS == ts {
			deleted = true
			continue
		}
		msgs = append(msgs, m)
	}
	ch.Messages = msgs

	return deleted
}

// ServeHTTP serves the Slack API.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s", req.Method, req.URL.Path)
	s.handler.ServeHTTP(w, req)
}
