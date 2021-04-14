package mockslack

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
)

//go:embed login.html
var loginPageHTML string

var loginPage = template.Must(
	template.New("login").
		Funcs(template.FuncMap{"StringsJoin": strings.Join}).
		Parse(loginPageHTML))

func (s *Server) loginMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/api") || ContextToken(req.Context()) != nil {
			next(w, req)
			return
		}

		if req.FormValue("action") == "login" {
			userID := req.FormValue("userID")

			if userID == "new" {
				usr := s.newUser(User{Name: req.FormValue("newUserName")})
				userID = usr.ID
			}

			tok := s.newToken(AuthToken{
				User:   userID,
				Scopes: []string{"user"},
			})

			http.SetCookie(w, &http.Cookie{
				Name:  TokenCookieName,
				Value: tok.ID,
				Path:  "/",
			})

			next(w, req.WithContext(WithToken(req.Context(), tok)))
			return
		}

		var renderContext struct {
			Users []User
			Data  url.Values
		}

		renderContext.Data = req.Form

		// remove used fields, if they existed
		renderContext.Data.Del("userID")
		renderContext.Data.Del("newUserName")
		renderContext.Data.Del("action")

		// show login
		err := loginPage.Execute(w, renderContext)
		if err != nil {
			log.Println("ERROR:", err)
		}
	}
}
