package oidc

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jmespath/go-jmespath"
	"github.com/pkg/errors"
	"github.com/target/goalert/auth"
	"github.com/target/goalert/config"
	"github.com/target/goalert/util/log"
	"golang.org/x/oauth2"
)

var _ auth.IdentityProvider = &Provider{}

const nonceCookieName = "goalert_oidc_nonce"

var b64enc = base64.URLEncoding.WithPadding(base64.NoPadding)

// Provider implements the auth.IdentityProvider interface by acting as a relying-party
// to a standard OIDC server.
type Provider struct {
	cfg Config

	mx        sync.Mutex
	providers map[string]*oidc.Provider
}

func (p *Provider) provider(ctx context.Context) (*oidc.Provider, error) {
	cfg := config.FromContext(ctx)
	p.mx.Lock()
	defer p.mx.Unlock()

	provider, ok := p.providers[cfg.OIDC.IssuerURL]
	if ok {
		return provider, nil
	}

	// oidc keeps the context and uses it after auto-discover is complete.
	// Giving it context.Background is a workaround to allow fetching keys
	// after init.
	oidcCtx := log.FromContext(ctx).BackgroundContext()
	provider, err := oidc.NewProvider(oidcCtx, cfg.OIDC.IssuerURL)
	if err != nil {
		return nil, err
	}

	p.providers[cfg.OIDC.IssuerURL] = provider
	return provider, nil
}

func (p *Provider) oaConfig(ctx context.Context) (*oauth2.Config, *oidc.IDTokenVerifier, error) {
	provider, err := p.provider(ctx)
	if err != nil {
		return nil, nil, err
	}
	cfg := config.FromContext(ctx)
	scopes := cfg.OIDC.Scopes
	// "openid" is a required scope for OpenID Connect flows.
	if scopes == "" {
		scopes = "openid profile email"
	}

	return &oauth2.Config{
		ClientID:     cfg.OIDC.ClientID,
		ClientSecret: cfg.OIDC.ClientSecret,

		Endpoint: provider.Endpoint(),

		Scopes: strings.Split(scopes, " "),
	}, provider.Verifier(&oidc.Config{ClientID: cfg.OIDC.ClientID}), nil
}

// NewProvider prepares a new Provider with the given config.
func NewProvider(ctx context.Context, cfg Config) (*Provider, error) {
	if cfg.Keyring == nil {
		return nil, errors.New("Keyring missing")
	}
	if cfg.NonceStore == nil {
		return nil, errors.New("NonceStore missing")
	}

	return &Provider{
		cfg:       cfg,
		providers: make(map[string]*oidc.Provider),
	}, nil
}

// Info returns the appropriate auth.ProviderInfo based on configuration.
//
// As OIDC requires no user input, only the Title is provided.
func (p *Provider) Info(ctx context.Context) auth.ProviderInfo {
	cfg := config.FromContext(ctx)
	title := "OIDC"
	if cfg.OIDC.OverrideName != "" {
		title = cfg.OIDC.OverrideName
	}
	return auth.ProviderInfo{
		Title:   title,
		Enabled: cfg.OIDC.Enable,
	}
}

func (p *Provider) newStateToken(nonceBytes []byte) (state string, err error) {
	buf := bytes.NewBuffer(nil)

	buf.Write(nonceBytes[:])
	buf.WriteByte('N')
	if err := binary.Write(buf, binary.BigEndian, time.Now().Unix()); err != nil {
		return "", err
	}

	sig, err := p.cfg.Keyring.Sign(buf.Bytes())
	if err != nil {
		return "", err
	}
	buf.Write(sig)

	// skip nonce for state token
	buf.Next(len(nonceBytes))

	return b64enc.EncodeToString(buf.Bytes()), nil
}

func (p *Provider) validateStateToken(ctx context.Context, nonce []byte, state string) (bool, error) {
	var buf bytes.Buffer
	buf.Write(nonce[:])
	data, err := b64enc.DecodeString(state)
	if err != nil {
		return false, err
	}
	buf.Write(data)
	data = buf.Bytes()
	if len(data) < 25 {
		return false, nil
	}
	valid, _ := p.cfg.Keyring.Verify(data[:25], data[25:])
	if !valid {
		return false, nil
	}
	if data[16] != 'N' {
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

	return true, nil
}

type claimsData struct {
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Email      string `json:"email"`
	Verified   bool   `json:"email_verified"`
}

// ExtractIdentity will return a redirect error for new auth requests, and provide a users identity
// for callback requests.
func (p *Provider) ExtractIdentity(route *auth.RouteInfo, w http.ResponseWriter, req *http.Request) (*auth.Identity, error) {
	ctx := req.Context()
	cfg := config.FromContext(ctx)

	name := "OIDC"
	if cfg.OIDC.OverrideName != "" {
		name = cfg.OIDC.OverrideName
	}

	switch route.RelativePath {
	case "/":
		nonce := p.cfg.NonceStore.New()
		stateToken, err := p.newStateToken(nonce[:])
		if err != nil {
			log.Log(req.Context(), errors.Wrap(err, "generate new state token"))
			return nil, auth.Error("Failed to generate state token.")
		}
		nonceStr := b64enc.EncodeToString(nonce[:])
		auth.SetCookie(w, req, nonceCookieName, nonceStr)

		oaCfg, _, err := p.oaConfig(ctx)
		if err != nil {
			return nil, err
		}
		oaCfg.RedirectURL = route.CurrentURL + "/callback"

		u := oaCfg.AuthCodeURL(stateToken, oidc.Nonce(nonceStr))
		return nil, auth.RedirectURL(u)
	case "/callback":
		// handled below
	default:
		return nil, auth.Error(fmt.Sprintf("Could not login due to wrong configuration for %s.", name))
	}

	stateToken := req.FormValue("state")
	nonceC, err := req.Cookie(nonceCookieName)
	if err != nil {
		return nil, auth.Error("There was a problem recognizing this browser. You can try again")
	}
	auth.ClearCookie(w, req, nonceCookieName)

	nonce, err := b64enc.DecodeString(nonceC.Value)
	if err != nil || len(nonce) != 16 {
		// We can't guarantee the current browser is the one we sent for auth (CSRF/XSS potential)
		return nil, auth.Error("There was a problem verifying this browser. You can try again")
	}
	valid, err := p.validateStateToken(req.Context(), nonce, stateToken)
	if err != nil {
		log.Log(req.Context(), errors.Wrap(err, "validate state token"))
		return nil, auth.Error("There was a redirection problem. You can try again")
	}
	if !valid {
		return nil, auth.Error("There was a problem while checking the request. You can try again")
	}

	oaCfg, verifier, err := p.oaConfig(ctx)
	if err != nil {
		return nil, err
	}
	oaCfg.RedirectURL = route.CurrentURL

	oauth2Token, err := oaCfg.Exchange(ctx, req.URL.Query().Get("code"))
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "exchange OIDC token"))
		return nil, auth.Error(fmt.Sprintf("Could not communicate with %s server. You can try again", name))
	}

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		log.Log(ctx, errors.New("id_token missing"))
		return nil, auth.Error(fmt.Sprintf("Bad response from %s server.", name))
	}

	// Parse and verify ID Token payload.
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "validate id_token"))
		return nil, auth.Error(fmt.Sprintf("Invalid response from %s server.", name))
	}

	remoteNonce, err := b64enc.DecodeString(idToken.Nonce)
	if err != nil || len(remoteNonce) != 16 || !bytes.Equal(remoteNonce, nonce) {
		return nil, auth.Error(fmt.Sprintf("Invalid nonce from %s server.", name))
	}
	var remoteNonceBytes [16]byte
	copy(remoteNonceBytes[:], remoteNonce)

	ok, err = p.cfg.NonceStore.Consume(ctx, remoteNonceBytes)
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "consume nonce value"))
		return nil, auth.Error("Could not login. You can try again")
	}
	if !ok {
		return nil, auth.Error("Could not login. You can try again")
	}

	// Extract custom claims
	var claims claimsData
	if err := idToken.Claims(&claims); err != nil {
		log.Log(ctx, errors.Wrap(err, "parse claims"))
		return nil, auth.Error(fmt.Sprintf("Invalid response from %s server.", name))
	}

	if claims.Name == "" {
		// We *should* always get name with the profile scope, but fall back to joining the given and family names
		// for misbehaving servers.
		claims.Name = strings.TrimSpace(claims.GivenName + " " + claims.FamilyName)
	}

	id := auth.Identity{
		Email:         claims.Email,
		Name:          claims.Name,
		EmailVerified: claims.Verified,
		SubjectID:     idToken.Subject,
	}

	var info interface{}
	getInfo := func(name, search string) interface{} {
		if err != nil {
			return nil
		}
		if info == nil {
			info, err = p.userInfoData(ctx, oaCfg.TokenSource(ctx, oauth2Token))
		}
		if err != nil {
			log.Log(ctx, err)
			return nil
		}
		res, searchErr := jmespath.Search(search, info)
		if searchErr != nil {
			log.Log(ctx, errors.Wrapf(searchErr, "lookup %s in UserInfo", name))
			return nil
		}

		return res
	}
	infoFieldStr := func(name, search string, field *string) {
		if search == "" {
			return
		}
		*field = ""
		res := getInfo(name, search)
		if res == nil {
			return
		}
		s, ok := res.(string)
		if !ok {
			log.Log(ctx, errors.Errorf("expected %s to be a string in UserInfo", name))
			return
		}
		*field = s
	}
	infoFieldBool := func(name, search string, field *bool) {
		if search == "" {
			return
		}
		*field = false
		res := getInfo(name, search)
		if res == nil {
			return
		}
		v, ok := res.(bool)
		if !ok {
			log.Log(ctx, errors.Errorf("expected %s to be a bool in UserInfo", name))
			return
		}
		*field = v
	}
	infoFieldStr("Email", cfg.OIDC.UserInfoEmailPath, &id.Email)
	infoFieldBool("EmailVerified", cfg.OIDC.UserInfoEmailVerifiedPath, &id.EmailVerified)
	infoFieldStr("Name", cfg.OIDC.UserInfoNamePath, &id.Name)

	return &id, nil
}

func (p *Provider) userInfoData(ctx context.Context, token oauth2.TokenSource) (interface{}, error) {
	provider, err := p.provider(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get provider")
	}

	info, err := provider.UserInfo(ctx, token)
	if err != nil {
		return nil, errors.Wrap(err, "fetch UserInfo")
	}

	var rawInfo interface{}
	if err := info.Claims(&rawInfo); err != nil {
		return nil, errors.Wrap(err, "parse UserInfo")
	}

	return rawInfo, nil
}
