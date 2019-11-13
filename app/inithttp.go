package app

import (
	"context"
	"net/http"
	"strings"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/target/goalert/config"
	"github.com/target/goalert/genericapi"
	"github.com/target/goalert/grafana"
	"github.com/target/goalert/mailgun"
	"github.com/target/goalert/notification/twilio"
	"github.com/target/goalert/site24x7"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/web"
	"go.opencensus.io/plugin/ochttp"
)

func (app *App) initHTTP(ctx context.Context) error {
	var traceMiddleware func(next http.Handler) http.Handler
	if app.cfg.StackdriverProjectID != "" {
		traceMiddleware = func(next http.Handler) http.Handler {
			return &ochttp.Handler{
				IsPublicEndpoint: true,
				Propagation:      &propagation.HTTPFormat{},
				Handler:          next,
			}
		}
	} else {
		traceMiddleware = func(next http.Handler) http.Handler {
			return &ochttp.Handler{
				IsPublicEndpoint: true,
				Handler:          next,
			}
		}
	}

	middleware := []func(http.Handler) http.Handler{
		traceMiddleware,
		// add app config to request context
		func(next http.Handler) http.Handler { return config.Handler(next, app.ConfigStore) },

		// request cooldown tracking (for graceful shutdown)
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if !strings.HasPrefix(req.URL.Path, "/health") {
					app.cooldown.Trigger()
				}
				next.ServeHTTP(w, req)
			})
		},

		config.ShortURLMiddleware,

		// redirect http to https if public URL is https
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				fwdProto := req.Header.Get("x-forwarded-proto")
				if fwdProto != "" {
					req.URL.Scheme = fwdProto
				} else if req.URL.Scheme == "" {
					if req.TLS == nil {
						req.URL.Scheme = "http"
					} else {
						req.URL.Scheme = "https"
					}
				}

				req.URL.Host = req.Host
				cfg := config.FromContext(req.Context())

				if app.cfg.DisableHTTPSRedirect || cfg.ValidReferer(req.URL.String(), req.URL.String()) {
					next.ServeHTTP(w, req)
					return
				}
				u := *req.URL
				u.Scheme = "https"
				if cfg.ValidReferer(req.URL.String(), u.String()) {
					http.Redirect(w, req, u.String(), http.StatusTemporaryRedirect)
					return
				}

				next.ServeHTTP(w, req)
			})
		},

		// limit auth check counts (fail-safe for loops or DB access)
		authCheckLimit(100),

		// request logging
		logRequest(app.cfg.LogRequests),

		// max request time
		timeout(2 * time.Minute),

		// remove public URL prefix
		stripPrefixMiddleware(),

		// limit max request size
		maxBodySizeMiddleware(app.cfg.MaxReqBodyBytes),

		// pause has to become before anything that uses the DB (like auth)
		app.pauseHandler,

		// authenticate requests
		app.authHandler.WrapHandler,

		// add auth info to request logs
		logRequestAuth,

		conReqLimit{
			perIntKey:  1,
			perService: 2,
			perUser:    3,
		}.Middleware,

		wrapGzip,
	}

	if app.cfg.Verbose {
		middleware = append(middleware, func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				next.ServeHTTP(w, req.WithContext(log.EnableDebug(req.Context())))
			})
		})
	}

	mux := http.NewServeMux()

	generic := genericapi.NewHandler(genericapi.Config{
		AlertStore:          app.AlertStore,
		IntegrationKeyStore: app.IntegrationKeyStore,
		HeartbeatStore:      app.HeartbeatStore,
		UserStore:           app.UserStore,
	})

	mux.Handle("/api/graphql", app.graphql2.Handler())
	mux.Handle("/api/graphql/explore", app.graphql2.PlayHandler())

	mux.HandleFunc("/api/v2/config", app.ConfigStore.ServeConfig)

	mux.HandleFunc("/api/v2/identity/providers", app.authHandler.ServeProviders)
	mux.HandleFunc("/api/v2/identity/logout", app.authHandler.ServeLogout)

	basicAuth := app.authHandler.IdentityProviderHandler("basic")
	mux.HandleFunc("/api/v2/identity/providers/basic", basicAuth)

	githubAuth := app.authHandler.IdentityProviderHandler("github")
	mux.HandleFunc("/api/v2/identity/providers/github", githubAuth)
	mux.HandleFunc("/api/v2/identity/providers/github/callback", githubAuth)

	oidcAuth := app.authHandler.IdentityProviderHandler("oidc")
	mux.HandleFunc("/api/v2/identity/providers/oidc", oidcAuth)
	mux.HandleFunc("/api/v2/identity/providers/oidc/callback", oidcAuth)

	mux.HandleFunc("/api/v2/mailgun/incoming", mailgun.IngressWebhooks(app.AlertStore, app.IntegrationKeyStore))
	mux.HandleFunc("/api/v2/grafana/incoming", grafana.GrafanaToEventsAPI(app.AlertStore, app.IntegrationKeyStore))
	mux.HandleFunc("/api/v2/site24x7/incoming", site24x7.Site24x7ToEventsAPI(app.AlertStore, app.IntegrationKeyStore))

	mux.HandleFunc("/api/v2/generic/incoming", generic.ServeCreateAlert)
	mux.HandleFunc("/api/v2/heartbeat/", generic.ServeHeartbeatCheck)
	mux.HandleFunc("/api/v2/user-avatar/", generic.ServeUserAvatar)

	mux.HandleFunc("/api/v2/twilio/message", app.twilioSMS.ServeMessage)
	mux.HandleFunc("/api/v2/twilio/message/status", app.twilioSMS.ServeStatusCallback)
	mux.HandleFunc("/api/v2/twilio/call", app.twilioVoice.ServeCall)
	mux.HandleFunc("/api/v2/twilio/call/status", app.twilioVoice.ServeStatusCallback)

	// Legacy (v1) API mappings
	mux.HandleFunc("/v1/graphql", app.graphql.ServeHTTP)
	muxRewrite(mux, "/v1/graphql2", "/api/graphql")
	muxRedirect(mux, "/v1/graphql2/explore", "/api/graphql/explore")
	muxRewrite(mux, "/v1/config", "/api/v2/config")
	muxRewrite(mux, "/v1/identity/providers", "/api/v2/identity/providers")
	muxRewritePrefix(mux, "/v1/identity/providers/", "/api/v2/identity/providers/")
	muxRewrite(mux, "/v1/identity/logout", "/api/v2/identity/logout")

	muxRewrite(mux, "/v1/webhooks/mailgun", "/api/v2/mailgun/incoming")
	muxRewrite(mux, "/v1/webhooks/grafana", "/api/v2/grafana/incoming")
	muxRewrite(mux, "/v1/api/alerts", "/api/v2/generic/incoming")
	muxRewritePrefix(mux, "/v1/api/heartbeat/", "/api/v2/heartbeat/")
	muxRewriteWith(mux, "/v1/api/users/", func(req *http.Request) *http.Request {
		parts := strings.Split(strings.TrimSuffix(req.URL.Path, "/avatar"), "/")
		req.URL.Path = "/api/v2/user-avatar/" + parts[len(parts)-1]
		return req
	})

	muxRewrite(mux, "/v1/twilio/sms/messages", "/api/v2/twilio/message")
	muxRewrite(mux, "/v1/twilio/sms/status", "/api/v2/twilio/message/status")
	muxRewrite(mux, "/v1/twilio/voice/call", "/api/v2/twilio/call?type=alert")
	muxRewrite(mux, "/v1/twilio/voice/alert-status", "/api/v2/twilio/call?type=alert-status")
	muxRewrite(mux, "/v1/twilio/voice/test", "/api/v2/twilio/call?type=test")
	muxRewrite(mux, "/v1/twilio/voice/stop", "/api/v2/twilio/call?type=stop")
	muxRewrite(mux, "/v1/twilio/voice/verify", "/api/v2/twilio/call?type=verify")
	muxRewrite(mux, "/v1/twilio/voice/status", "/api/v2/twilio/call/status")

	twilioHandler := twilio.WrapValidation(
		// go back to the regular mux after validation
		twilio.WrapHeaderHack(mux),
		*app.twilioConfig,
	)

	topMux := http.NewServeMux()

	// twilio calls should go through the validation handler first
	// since the signature is based on the original URL
	topMux.Handle("/v1/twilio/", twilioHandler)
	topMux.Handle("/api/v2/twilio/", twilioHandler)

	topMux.Handle("/v1/", mux)
	topMux.Handle("/api/", mux)

	topMux.HandleFunc("/health", app.healthCheck)
	topMux.HandleFunc("/health/engine", app.engineStatus)

	webH, err := web.NewHandler(app.cfg.UIURL)
	if err != nil {
		return err
	}
	// non-API/404s go to UI handler
	topMux.Handle("/", webH)

	app.srv = &http.Server{
		Handler: applyMiddleware(topMux, middleware...),

		ReadHeaderTimeout: time.Second * 30,
		ReadTimeout:       time.Minute,
		WriteTimeout:      time.Minute,
		IdleTimeout:       time.Minute * 2,
		MaxHeaderBytes:    app.cfg.MaxReqHeaderBytes,
	}

	// Ingress/load balancer/proxy can do keep-alives, backend doesn't need it.
	// It also makes zero downtime deploys nearly impossible; an idle connection
	// could have an in-flight request when the server closes it.
	app.srv.SetKeepAlivesEnabled(false)

	return nil
}
