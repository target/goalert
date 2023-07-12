package github

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v53/github"
	"github.com/pkg/errors"
	"github.com/target/goalert/auth"
	"github.com/target/goalert/config"
	"github.com/target/goalert/util/log"
	"golang.org/x/oauth2"
)

const stateCookieName = "goalert_github_auth_state"

// Info implements the auth.Provider interface.
func (Provider) Info(ctx context.Context) auth.ProviderInfo {
	cfg := config.FromContext(ctx)
	return auth.ProviderInfo{
		Title:   "GitHub",
		Enabled: cfg.GitHub.Enable,
	}
}

func (p *Provider) newStateToken() (string, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('N')

	tok := p.c.NonceStore.New()
	buf.Write(tok[:])

	if err := binary.Write(buf, binary.BigEndian, time.Now().Unix()); err != nil {
		return "", err
	}

	sig, err := p.c.Keyring.Sign(buf.Bytes())
	if err != nil {
		return "", err
	}
	buf.Write(sig)

	return base64.URLEncoding.EncodeToString(buf.Bytes()), nil
}

func (p *Provider) validateStateToken(ctx context.Context, s string) (bool, error) {
	data, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return false, err
	}
	if len(data) < 25 {
		return false, nil
	}
	valid, _ := p.c.Keyring.Verify(data[:25], data[25:])
	if !valid {
		return false, nil
	}
	if data[0] != 'N' {
		return false, nil
	}
	var id [16]byte
	copy(id[:], data[1:])

	unix := int64(binary.BigEndian.Uint64(data[17:]))
	t := time.Unix(unix, 0)
	if time.Since(t) > time.Hour {
		return false, nil
	}
	if time.Until(t) > time.Minute*5 {
		// too far in the future (clock drift)
		return false, nil
	}

	return p.c.NonceStore.Consume(ctx, id)
}

// ExtractIdentity implements the auth.IdentityProvider interface handling both auth and callback endpoints.
func (p *Provider) ExtractIdentity(route *auth.RouteInfo, w http.ResponseWriter, req *http.Request) (*auth.Identity, error) {
	ctx := req.Context()
	cfg := config.FromContext(ctx)

	switch route.RelativePath {
	case "/":
		tok, err := p.newStateToken()
		if err != nil {
			log.Log(req.Context(), errors.Wrap(err, "generate new state token"))
			return nil, auth.Error("Failed to generate state token.")
		}

		auth.SetCookie(w, req, stateCookieName, tok, false)
		u := authConfig(ctx).AuthCodeURL(tok, oauth2.ApprovalForce)

		return nil, auth.RedirectURL(u)
	case "/callback":
		// handled below
	default:
		return nil, auth.Error("Invalid callback URL specified in GitHub application config.")
	}

	tokStr := req.FormValue("state")
	stateCookie, err := req.Cookie("goalert_github_auth_state")
	if err != nil || stateCookie.Value != tokStr {
		return nil, auth.Error("Invalid state token.")
	}
	auth.ClearCookie(w, req, stateCookieName, false)

	valid, err := p.validateStateToken(req.Context(), tokStr)
	if err != nil {
		log.Log(req.Context(), errors.Wrap(err, "validate state token"))
		return nil, auth.Error("Could not validate state token.")
	}
	if !valid {
		return nil, auth.Error("Invalid state token.")
	}

	errorDesc := req.FormValue("error_description")
	if errorDesc != "" {
		return nil, auth.Error(errorDesc)
	}

	oaCfg := authConfig(ctx)

	tok, err := oaCfg.Exchange(ctx, req.FormValue("code"))
	if err != nil {
		log.Log(ctx, fmt.Errorf("github: exchange token: %w", err))
		return nil, auth.Error("Failed to get token from GitHub.")
	}

	if !tok.Valid() {
		return nil, auth.Error("Invalid token returned from GitHub.")
	}

	c := oaCfg.Client(ctx, tok)
	g := github.NewClient(c)
	if cfg.GitHub.EnterpriseURL != "" {
		g.BaseURL, err = url.Parse(strings.TrimSuffix(cfg.GitHub.EnterpriseURL, "/") + "/api/v3/")
		if err != nil {
			return nil, err
		}
	}

	u, _, err := g.Users.Get(ctx, "")
	if err != nil {
		log.Log(ctx, fmt.Errorf("github: fetch user: %w", err))
		return nil, auth.Error("Failed to fetch user profile from GitHub.")
	}

	var inUsers bool
	login := strings.ToLower(u.GetLogin())
	ctx = log.WithFields(ctx, log.Fields{
		"github_id":    u.GetID(),
		"github_login": u.GetLogin(),
		"github_name":  u.GetName(),
	})

	for _, u := range cfg.GitHub.AllowedUsers {
		if u == "*" || login == strings.ToLower(u) {
			inUsers = true
			break
		}
	}

	var inOrg bool
	if !inUsers && len(cfg.GitHub.AllowedOrgs) > 0 {
		for _, o := range cfg.GitHub.AllowedOrgs {
			if strings.Contains(o, "/") {
				// skip teams (process below)
				continue
			}
			m, _, err := g.Organizations.IsMember(ctx, o, login)
			if err != nil {
				log.Log(ctx, errors.Wrap(err, "fetch GitHub org membership"))
				return nil, auth.Error("Failed to read GitHub org membership")
			}
			if m {
				inOrg = true
				ctx = log.WithField(ctx, "github_org", o)
				log.Debugf(ctx, "GitHub Auth matched org")
				break
			}
		}

		if !inOrg {

			opt := &github.ListOptions{}
			teams := make([]string, 0, 30)

		CheckTeams:
			for {
				tm, resp, err := g.Teams.ListUserTeams(ctx, opt)
				if err != nil {
					log.Log(ctx, errors.Wrap(err, "fetch GitHub teams"))
					return nil, auth.Error("Failed to read GitHub team membership")
				}
				for _, t := range tm {
					teamName := strings.ToLower(t.Organization.GetLogin()) + "/" + strings.ToLower(t.GetSlug())
					teams = append(teams, teamName)
					if containsOrg(cfg.GitHub.AllowedOrgs, teamName) {
						inOrg = true
						ctx = log.WithField(ctx, "github_team", teamName)
						log.Debugf(ctx, "GitHub Auth matched team")
						break CheckTeams
					}
				}
				if resp.NextPage == 0 {
					break
				}
				opt.Page = resp.NextPage
			}

			// if still no match, log everything
			if !inOrg {
				log.Debugf(log.WithFields(ctx, log.Fields{
					"AllowedOrgs":     cfg.GitHub.AllowedOrgs,
					"TeamMemberships": teams,
				}), "not in any matching team or org")
			}
		}
	}

	if !inUsers && !inOrg {
		return nil, auth.Error("Not a member of an allowed org or whitelisted user.")
	}
	if strings.TrimSpace(u.GetName()) == "" {
		return nil, auth.Error("GitHub user has no display name set.")
	}

	return &auth.Identity{
		Email:     u.GetEmail(),
		Name:      u.GetName(),
		SubjectID: strconv.FormatInt(u.GetID(), 10),
	}, nil
}
