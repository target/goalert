package mockslack

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var authPage = template.Must(
	template.New("authorize").
		Funcs(template.FuncMap{"StringsJoin": strings.Join}).
		Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta http-equiv="X-UA-Compatible" content="ie=edge">
	<title>Mock Slack - Authorize</title>
</head>
<body>
	<center>
	<h1>Authorize</h1>
	<h2>Logged in as: {{.UserName}}</h2>
	<h3>Allow the application {{.AppName}} access to the following scopes:</h3>
	<ul>
		{{range .Scopes}}
			<li>{{.}}</li>
		{{end}}
	</ul>
	<hr>
	<form method="POST">
		{{- range $key, $value := .Data}}
		<input name={{ $key }} type="hidden" value={{StringsJoin $value " "}} />
		{{- end}}

		<input id="action" type="hidden" name="action" value="confirm" />
		<button type="submit" style="color:gray;width:20%;height:2em;font-size: 3em" onclick="document.getElementById('action').setAttribute('value', 'cancel')">Cancel</button>
		<button type="submit" style="background-color: green;width:20%;height:2em;font-size: 3em">Authorize</button>
	</form>
	
	</center>
</body>
</html>
`))

func (s *Server) ServeOAuthAuthorize(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	if respondErr(w, checkPermission(ctx, "user")) {
		return
	}

	clientID := req.FormValue("client_id")
	var renderData struct {
		AppName  string
		UserName string
		Scopes   []string
		Data     url.Values
	}
	renderData.Data = req.Form

	redir, err := url.Parse(req.FormValue("redirect_uri"))
	if err != nil {
		respondWith(w, &response{Err: "bad_redirect_uri"})
		return
	}

	errResp := func(msg string) {
		q := redir.Query()
		q.Set("state", req.FormValue("state"))
		q.Set("error", msg)
		redir.RawQuery = q.Encode()
		http.Redirect(w, req, redir.String(), http.StatusFound)
	}

	app := s.app(clientID)
	if app == nil {
		errResp("invalid_client_id")
		return
	}
	renderData.AppName = app.Name

	if req.FormValue("action") == "cancel" {
		errResp("access_denied")
		return
	}

	uid := userID(ctx)
	renderData.UserName = s.user(uid).Name
	scopes := strings.Split(req.FormValue("scope"), " ")
	renderData.Scopes = scopes
	if req.FormValue("action") != "confirm" {
		err = authPage.Execute(w, renderData)
		if err != nil {
			log.Println("ERROR:", err)
		}
		return
	}

	code := s.addUserAppScope(uid, clientID, scopes...)

	q := redir.Query()
	q.Del("error")
	q.Set("code", code)
	q.Set("state", req.FormValue("state"))
	redir.RawQuery = q.Encode()
	http.Redirect(w, req, redir.String(), http.StatusFound)
}
