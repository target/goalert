package mockslack

type User struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	TeamID string `json:"team_id"`
}
type userState struct {
	User
	appTokens map[string]*AuthToken
}

func (st *state) user(id string) *userState {
	st.mx.Lock()
	defer st.mx.Unlock()

	return st.users[id]
}
func (st *state) newUser(u User) User {
	st.mx.Lock()
	defer st.mx.Unlock()

	if u.ID == "" {
		u.ID = st.gen.UserID()
	}
	if u.TeamID == "" {
		u.TeamID = st.teamID
	}
	st.users[u.ID] = &userState{User: u, appTokens: make(map[string]*AuthToken)}

	return u
}

func (st *state) addUserAppScope(userID, clientID string, scopes ...string) string {
	st.mx.Lock()
	defer st.mx.Unlock()

	if st.users[userID].appTokens[clientID] == nil {
		tok := &AuthToken{ID: st.gen.UserAccessToken(), User: userID, Scopes: scopes}
		st.tokens[tok.ID] = tok
		st.users[userID].appTokens[clientID] = tok

		code := st.gen.TokenCode()
		st.tokenCodes[code] = &tokenCode{AuthToken: tok, ClientID: clientID}
		return code
	}

	tok := st.users[userID].appTokens[clientID]

	for _, scope := range scopes {
		if !contains(tok.Scopes, scope) {
			tok.Scopes = append(tok.Scopes, scope)
		}
	}

	code := st.gen.TokenCode()
	st.tokenCodes[code] = &tokenCode{AuthToken: tok, ClientID: clientID}
	return code
}
