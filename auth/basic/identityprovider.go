package basic

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"github.com/target/goalert/auth"
	"github.com/target/goalert/config"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation/validate"
)

// Info implements the auth.Provider interface.
func (Provider) Info(ctx context.Context) auth.ProviderInfo {
	cfg := config.FromContext(ctx)
	return auth.ProviderInfo{
		Title: "Basic",
		Fields: []auth.Field{
			{ID: "username", Label: "Username", Required: true},
			{ID: "password", Label: "Password", Password: true, Required: true},
		},
		Enabled: !cfg.Auth.DisableBasic,
	}
}

func userPass(req *http.Request) (string, string) {
	if req.URL.User == nil {
		return req.FormValue("username"), req.FormValue("password")
	}

	p, _ := req.URL.User.Password()
	return req.URL.User.Username(), p
}

// ExtractIdentity implements the auth.IdentityProvider interface, providing identity based
// on the given username and password fields.
func (p *Provider) ExtractIdentity(route *auth.RouteInfo, w http.ResponseWriter, req *http.Request) (*auth.Identity, error) {
	ctx := req.Context()

	username, password := userPass(req)
	err := validate.Username("Username", username)
	if err != nil {
		return nil, auth.Error("invalid username")
	}
	ctx = log.WithField(ctx, "username", username)

	err = p.lim.Lock(ctx, username)
	if errutil.HTTPError(ctx, w, err) {
		return nil, err
	}
	defer p.lim.Unlock(username)

	_, err = p.b.Validate(ctx, username, password)
	if err != nil {
		log.Debug(ctx, errors.Wrap(err, "basic login"))
		auth.Delay(ctx)
		return nil, auth.Error("unknown username/password")
	}

	return &auth.Identity{
		SubjectID: username,
	}, nil
}
