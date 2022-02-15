package graphqlapp

import (
	_ "embed"
	"html/template"
	"net/http"

	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/errutil"
)

const graphiqlVersion = "1.5.19"
const reactVersion = "17.0.2"

//go:embed playground.html
var playHTML string

var playTmpl = template.Must(template.New("graphqlPlayground").Parse(playHTML))

type webResource struct {
	Version string
	SRI     string
}

func (a *App) PlayHandler(w http.ResponseWriter, req *http.Request) {
	var data struct {
		ApplicationName string
		GraphiqlJS      webResource
		GraphiqlCSS     webResource
		React           webResource
		ReactDOM        webResource
	}

	ctx := req.Context()

	err := permission.LimitCheckAny(ctx)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	cfg := config.FromContext(ctx)

	data.ApplicationName = cfg.ApplicationName()
	data.GraphiqlJS = webResource{
		Version: graphiqlVersion,
		SRI:     "sha256-xrkMi9ErcGZKBQCFEhvMQva3GUX6QE3GPr4iI4BD/ws=",
	}
	data.GraphiqlCSS = webResource{
		Version: graphiqlVersion,
		SRI:     "sha256-HADQowUuFum02+Ckkv5Yu5ygRoLllHZqg0TFZXY7NHI=",
	}
	data.React = webResource{
		Version: reactVersion,
		SRI:     "sha256-Ipu/TQ50iCCVZBUsZyNJfxrDk0E2yhaEIz0vqI+kFG8=",
	}
	data.ReactDOM = webResource{
		Version: reactVersion,
		SRI:     "sha256-nbMykgB6tsOFJ7OdVmPpdqMFVk4ZsqWocT6issAPUF0=",
	}

	err = playTmpl.Execute(w, data)
	if errutil.HTTPError(ctx, w, err) {
		return
	}
}
