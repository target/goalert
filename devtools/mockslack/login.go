package mockslack

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var loginPage = template.Must(
	template.New("login").
		Funcs(template.FuncMap{"StringsJoin": strings.Join}).
		Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta http-equiv="X-UA-Compatible" content="ie=edge">
	<title>Mock Slack - Login</title>
</head>
<body>
	<center>
	<h1>Login</h1>
	<h3>Select an existing user, or create a new one.</h3>
  <hr>
  <script>
    function newUserChange(e) {
		if (!e.target.checked) {
			document.getElementById('newUserName').setAttribute('disabled', 'disabled')
		} else {
			document.getElementById('newUserName').removeAttribute('disabled')
		}
      
    }
  </script>
	<form method="POST">
		{{- range $key, $value := .Data}}
		<input name={{ $key }} type="hidden" value={{StringsJoin $value " "}} />
		{{- end}}

    {{range .Users}}
    <label><input type="radio" name="userID" value="{{.ID}}" />{{.Name}}</label><br>
	{{end}}
	<br>
    <label><input type="radio" name="userID" value="new" onchange="newUserChange" />+ Create New User</label>
	<br>
    <label>New User Name: <input id="newUserName" type="text" name="newUserName"  /></label>

		<input id="action" type="hidden" name="action" value="login" />
		<hr>
		<button type="submit" style="color:gray;width:20%;height:2em;font-size: 3em" onclick="document.getElementById('action').setAttribute('value', 'cancel')">Cancel</button>
		<button type="submit" style="background-color: green;width:20%;height:2em;font-size: 3em">Login</button>
	</form>

	</center>
</body>
</html>
`))

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
