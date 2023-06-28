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
