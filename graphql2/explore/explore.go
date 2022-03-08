package explore

import (
	"embed"
	_ "embed"
	"html/template"
	"net/http"

	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/errutil"
)

//go:embed explore.html
var htmlStr string

//go:embed build
var FS embed.FS

var playTmpl = template.Must(template.New("graphqlPlayground").Parse(htmlStr))

func Handler(w http.ResponseWriter, req *http.Request) {
	var data struct {
		ApplicationName string
	}

	ctx := req.Context()
	cfg := config.FromContext(ctx)
	err := permission.LimitCheckAny(ctx)

	if err != nil {
		// TODO redirect to login page instead of base path
		http.Redirect(w, req, "../../", http.StatusTemporaryRedirect)
		return
	}

	data.ApplicationName = cfg.ApplicationName()

	err = playTmpl.Execute(w, data)
	if errutil.HTTPError(ctx, w, err) {
		return
	}
}
