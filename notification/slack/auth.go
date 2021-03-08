package slack

import (
	"fmt"
	"net/http"
)

// 4. given a "code" field that expires after 10m
// 5. call oath.v2.access method with code
//   6. `curl -F code=1234 -F client_id=3336676.569200954261 -F client_secret=ABCDEFGH https://slack.com/api/oauth.v2.access`
// 7. token is returned under `authed_user.access_token`
// 8. store token in database with userID relation
// 9. redirect user to slack:// uri?
//
// notes:
// - oath tokens do not expire
// - provide a user_scope parameter with requested user scopes instead of, or in addition to, the scope parameter

func (h *Handler) ServeUserAuthCallback(w http.ResponseWriter, req *http.Request) {
	code := req.FormValue("code")
	fmt.Println("req: ", req)
	fmt.Println("code: ", code)
}
