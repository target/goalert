package app

import (
    "bytes"
    "context"
    "encoding/json"
    "database/sql"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
    "time"

    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/target/goalert/config"
	"github.com/target/goalert/expflag"
	"github.com/target/goalert/genericapi"
	"github.com/target/goalert/grafana"
	"github.com/target/goalert/mailgun"
	"github.com/target/goalert/notification/twilio"
	"github.com/target/goalert/permission"
	prometheus "github.com/target/goalert/prometheusalertmanager"
	"github.com/target/goalert/site24x7"
	"github.com/target/goalert/util/errutil"
    "github.com/target/goalert/util/log"
    "github.com/target/goalert/web"
    webpush "github.com/SherClockHolmes/webpush-go"
)

func (app *App) initHTTP(ctx context.Context) error {
	middleware := []func(http.Handler) http.Handler{
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				next.ServeHTTP(w, req.WithContext(app.Context(req.Context())))
			})
		},

		withSecureHeaders(app.cfg.EnableSecureHeaders, strings.HasPrefix(app.cfg.PublicURL, "https://")),

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

				u, err := url.ParseRequestURI(req.RequestURI)
				if errutil.HTTPError(req.Context(), w, err) {
					return
				}
				u.Scheme = "https"
				u.Host = req.Host
				if cfg.ValidReferer(req.URL.String(), u.String()) {
					http.Redirect(w, req, u.String(), http.StatusTemporaryRedirect)
					return
				}

				next.ServeHTTP(w, req)
			})
		},

		// limit external calls (fail-safe for loops or DB access)
		extCallLimit(100),

		// request logging
		logRequest(app.cfg.LogRequests),

		// max request time
		timeout(2 * time.Minute),

		func(next http.Handler) http.Handler {
			return http.StripPrefix(app.cfg.HTTPPrefix, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "" {
					req.URL.Path = "/"
				}

				next.ServeHTTP(w, req)
			}))
		},

		// limit max request size
		maxBodySizeMiddleware(app.cfg.MaxReqBodyBytes),

		// authenticate requests
		app.AuthHandler.WrapHandler,

		// add auth info to request logs
		logRequestAuth,

		LimitConcurrencyByAuthSource,

		wrapGzip,
	}

	if app.cfg.Verbose {
		middleware = append(middleware, func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				next.ServeHTTP(w, req.WithContext(log.WithDebug(req.Context())))
			})
		})
	}

    mux := http.NewServeMux()

    // Ensure table for persisted Web Push subscriptions (quick bootstrap without migration).
    if _, err := app.db.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS user_web_push_subscriptions (
            endpoint text PRIMARY KEY,
            user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            data jsonb NOT NULL,
            created_at timestamptz NOT NULL DEFAULT now()
        );
        CREATE INDEX IF NOT EXISTS user_web_push_subscriptions_user_id_idx ON user_web_push_subscriptions (user_id);
    `); err != nil {
        return err
    }
    sendPush := func(ctx context.Context, payload []byte, userIDs ...string) {
        // Use a background context carrying logger + config to avoid request cancellation.
        ctx = app.Context(log.FromContext(ctx).BackgroundContext())
        cfg := config.FromContext(ctx)
        if !cfg.WebPush.Enable || cfg.WebPush.VAPIDPublicKey == "" || cfg.WebPush.VAPIDPrivateKey == "" {
            log.Logf(ctx, "webpush: disabled or missing VAPID keys; skipping send (enabled=%t)", cfg.WebPush.Enable)
            return
        }
        type subKeys struct{ P256dh, Auth string }
        type sub struct{
            Endpoint string `json:"endpoint"`
            Keys     subKeys `json:"keys"`
        }
        var rows *sql.Rows
        var err error
        if len(userIDs) == 0 {
            rows, err = app.db.QueryContext(ctx, `select data from user_web_push_subscriptions`)
        } else {
            var sb strings.Builder
            sb.WriteString("select data from user_web_push_subscriptions where user_id in (")
            for i := range userIDs {
                if i > 0 { sb.WriteString(",") }
                fmt.Fprintf(&sb, "$%d::uuid", i+1)
            }
            sb.WriteString(")")
            args := make([]any, len(userIDs))
            for i, id := range userIDs { args[i] = id }
            rows, err = app.db.QueryContext(ctx, sb.String(), args...)
        }
        if err != nil {
            log.Logf(ctx, "webpush: query subscriptions failed: %v", err)
            return
        }
        defer rows.Close()
        var total int
        for rows.Next() {
            var raw json.RawMessage
            if err := rows.Scan(&raw); err != nil { continue }
            var s sub
            if err := json.Unmarshal(raw, &s); err != nil || s.Endpoint == "" || s.Keys.P256dh == "" || s.Keys.Auth == "" { continue }
            subObj := &webpush.Subscription{ Endpoint: s.Endpoint, Keys: webpush.Keys{ Auth: s.Keys.Auth, P256dh: s.Keys.P256dh } }
            go func(subObj *webpush.Subscription) {
                // Log destination provider based on endpoint host for diagnostics (iOS/APNs vs others)
                host := ""
                if u, perr := url.Parse(subObj.Endpoint); perr == nil { host = u.Host }
                provider := "unknown"
                if strings.Contains(host, "web.push.apple.com") { provider = "apple" }
                if strings.Contains(host, "fcm.googleapis.com") || strings.Contains(host, "firebase") { provider = "fcm" }
                if strings.Contains(host, "mozilla") || strings.Contains(host, "push.services.mozilla.com") { provider = "mozilla" }
                if strings.Contains(host, "notify.windows.com") { provider = "wns" }

                // Use a stable, valid URL as VAPID subject based on configured PublicURL.
                subj := cfg.PublicURL()
                if subj == "" {
                    subj = "mailto:no-reply@localhost"
                }

                resp, err := webpush.SendNotification(payload, subObj, &webpush.Options{
                    Subscriber:      subj,
                    VAPIDPublicKey:  cfg.WebPush.VAPIDPublicKey,
                    VAPIDPrivateKey: cfg.WebPush.VAPIDPrivateKey,
                    TTL:             60,
                    Urgency:         "high",
                })
                var status int
                var bodySnippet string
                if resp != nil {
                    status = resp.StatusCode
                    // Read small body for diagnostics on error statuses
                    if status >= 400 {
                        b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
                        bodySnippet = string(b)
                    }
                    _ = resp.Body.Close()
                }
                if err != nil {
                    log.Logf(ctx, "webpush: send failed; provider=%s host=%s status=%d err=%v", provider, host, status, err)
                    return
                }
                if status >= 400 {
                    log.Logf(ctx, "webpush: send non-2xx; provider=%s host=%s status=%d body=%q", provider, host, status, bodySnippet)
                } else {
                    log.Logf(ctx, "webpush: send complete; provider=%s host=%s status=%d", provider, host, status)
                }
                // Clean up expired/invalid subscriptions
                if status == http.StatusGone || status == http.StatusNotFound {
                    if _, derr := app.db.ExecContext(ctx, `delete from user_web_push_subscriptions where endpoint = $1`, subObj.Endpoint); derr != nil {
                        log.Logf(ctx, "webpush: cleanup failed; endpoint=%s err=%v", subObj.Endpoint, derr)
                    } else {
                        log.Logf(ctx, "webpush: removed expired subscription; endpoint=%s", subObj.Endpoint)
                    }
                }
            }(subObj)
            total++
        }
        log.Logf(ctx, "webpush: queued sends; targeted-users=%d total-subs=%d", len(userIDs), total)
    }

	generic := genericapi.NewHandler(genericapi.Config{
		AlertStore:          app.AlertStore,
		IntegrationKeyStore: app.IntegrationKeyStore,
		HeartbeatStore:      app.HeartbeatStore,
		UserStore:           app.UserStore,
	})

	mux.Handle("POST /api/graphql", app.graphql2.Handler())

	mux.HandleFunc("GET /api/v2/config", app.ConfigStore.ServeConfig)
	mux.HandleFunc("PUT /api/v2/config", app.ConfigStore.ServeConfig)

    // Accept Web Push subscription payloads (UI uses this endpoint).
    mux.HandleFunc("POST /api/push/subscribe", func(w http.ResponseWriter, req *http.Request) {
        data, err := io.ReadAll(io.LimitReader(req.Body, 1<<20))
        if err != nil { http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest); return }
        var tmp struct{ Endpoint string `json:"endpoint"` }
        _ = json.Unmarshal(data, &tmp)
        if tmp.Endpoint == "" { http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest); return }
        uid := permission.UserID(req.Context())
        if uid == "" { http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized); return }
        // Log UA and endpoint host for iOS/Safari diagnostics
        ua := req.Header.Get("User-Agent")
        host := ""
        if u, perr := url.Parse(tmp.Endpoint); perr == nil { host = u.Host }
        provider := "unknown"
        if strings.Contains(host, "web.push.apple.com") { provider = "apple" }
        if strings.Contains(host, "fcm.googleapis.com") || strings.Contains(host, "firebase") { provider = "fcm" }
        if strings.Contains(host, "mozilla") || strings.Contains(host, "push.services.mozilla.com") { provider = "mozilla" }
        if strings.Contains(host, "notify.windows.com") { provider = "wns" }
        log.Logf(req.Context(), "webpush: subscribe; user=%s provider=%s host=%s ua=%q", uid, provider, host, ua)
        _, err = app.db.ExecContext(req.Context(), `
            insert into user_web_push_subscriptions (endpoint, user_id, data)
            values ($1, $2::uuid, $3::jsonb)
            on conflict (endpoint) do update set user_id = excluded.user_id, data = excluded.data
        `, tmp.Endpoint, uid, data)
        if err != nil { http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError); return }
        w.WriteHeader(http.StatusNoContent)
    })

    // List current user's push subscriptions.
    mux.HandleFunc("GET /api/push/subscriptions", func(w http.ResponseWriter, req *http.Request) {
        uid := permission.UserID(req.Context())
        if uid == "" { http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized); return }
        rows, err := app.db.QueryContext(req.Context(), `
            select endpoint, created_at from user_web_push_subscriptions where user_id = $1::uuid order by created_at desc
        `, uid)
        if err != nil { http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError); return }
        defer rows.Close()
        type respT struct{ Endpoint, Host, Provider string; CreatedAt time.Time }
        var out []respT
        for rows.Next() {
            var endpoint string
            var created time.Time
            if err := rows.Scan(&endpoint, &created); err != nil { continue }
            h := ""; prov := "unknown"
            if u, e := url.Parse(endpoint); e == nil { h = u.Host }
            if strings.Contains(h, "web.push.apple.com") { prov = "apple" }
            if strings.Contains(h, "fcm.googleapis.com") || strings.Contains(h, "firebase") { prov = "fcm" }
            if strings.Contains(h, "mozilla") || strings.Contains(h, "push.services.mozilla.com") { prov = "mozilla" }
            if strings.Contains(h, "notify.windows.com") { prov = "wns" }
            out = append(out, respT{ Endpoint: endpoint, Host: h, Provider: prov, CreatedAt: created })
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(out)
    })

    // Delete a specific subscription for current user by endpoint.
    mux.HandleFunc("DELETE /api/push/subscriptions", func(w http.ResponseWriter, req *http.Request) {
        uid := permission.UserID(req.Context())
        if uid == "" { http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized); return }
        var endpoint string
        // Allow query param or JSON body
        endpoint = req.URL.Query().Get("endpoint")
        if endpoint == "" {
            var p struct{ Endpoint string `json:"endpoint"` }
            body, _ := io.ReadAll(io.LimitReader(req.Body, 1<<16))
            _ = json.Unmarshal(body, &p)
            endpoint = p.Endpoint
        }
        if endpoint == "" { http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest); return }
        // Restrict delete to current user's rows
        res, err := app.db.ExecContext(req.Context(), `
            delete from user_web_push_subscriptions where endpoint = $1 and user_id = $2::uuid
        `, endpoint, uid)
        if err != nil { http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError); return }
        n, _ := res.RowsAffected()
        log.Logf(req.Context(), "webpush: delete subscription; user=%s endpoint=%s removed=%d", uid, endpoint, n)
        w.WriteHeader(http.StatusNoContent)
    })

    // Admin/test endpoint to send a sample notification to all saved subscriptions.
    mux.HandleFunc("POST /admin/test/webpush", func(w http.ResponseWriter, req *http.Request) {
        if err := permission.LimitCheckAny(req.Context(), permission.Admin); err != nil {
            http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
            return
        }
        payload, _ := json.Marshal(struct{
            OnCall bool   `json:"onCall"`
            Title  string `json:"title"`
            Body   string `json:"body"`
            URL    string `json:"url"`
        }{ OnCall: true, Title: "Test Notification", Body: "This is a test", URL: "/alerts" })
        go sendPush(req.Context(), payload)
        w.WriteHeader(http.StatusNoContent)
    })

	mux.HandleFunc("GET /api/v2/identity/providers", app.AuthHandler.ServeProviders)
	mux.HandleFunc("POST /api/v2/identity/logout", app.AuthHandler.ServeLogout)

	basicAuth := app.AuthHandler.IdentityProviderHandler("basic")
	mux.HandleFunc("POST /api/v2/identity/providers/basic", basicAuth)

	githubAuth := app.AuthHandler.IdentityProviderHandler("github")
	mux.HandleFunc("POST /api/v2/identity/providers/github", githubAuth)
	mux.HandleFunc("GET /api/v2/identity/providers/github/callback", githubAuth)

	oidcAuth := app.AuthHandler.IdentityProviderHandler("oidc")
	mux.HandleFunc("POST /api/v2/identity/providers/oidc", oidcAuth)
	mux.HandleFunc("GET /api/v2/identity/providers/oidc/callback", oidcAuth)

	if expflag.ContextHas(ctx, expflag.UnivKeys) {
		mux.HandleFunc("POST /api/v2/uik", app.UIKHandler.ServeHTTP)
	}
	mux.HandleFunc("POST /api/v2/mailgun/incoming", mailgun.IngressWebhooks(app.AlertStore, app.IntegrationKeyStore))
{
    inner := grafana.GrafanaToEventsAPI(app.AlertStore, app.IntegrationKeyStore)
    mux.HandleFunc("POST /api/v2/grafana/incoming", func(w http.ResponseWriter, req *http.Request) {
        b, _ := io.ReadAll(io.LimitReader(req.Body, 1<<20))
        req.Body = io.NopCloser(bytes.NewReader(b))
        inner(w, req)

        var p struct{
            Title string `json:"title"`
            CommonAnnotations map[string]string `json:"commonAnnotations"`
        }
        _ = json.Unmarshal(b, &p)
        title := p.Title
        body := p.CommonAnnotations["summary"]
        if body == "" { body = p.CommonAnnotations["message"] }
        if title == "" && body == "" { return }
        // Target only on-call users for the service
        svcID := permission.ServiceID(req.Context())
        var tgtUserIDs []string
        if svcID != "" {
            permission.SudoContext(req.Context(), func(ctx context.Context) {
                if oc, err := app.OnCallStore.OnCallUsersByService(ctx, svcID); err == nil {
                    for _, u := range oc { tgtUserIDs = append(tgtUserIDs, u.UserID) }
                } else {
                    log.Logf(ctx, "webpush: grafana oncall lookup failed: %v", err)
                }
            })
        }
        log.Logf(req.Context(), "webpush: grafana incoming; serviceID=%s title=%q body.len=%d target-users=%d", svcID, title, len(body), len(tgtUserIDs))
        note := struct{
            OnCall bool   `json:"onCall"`
            Title  string `json:"title"`
            Body   string `json:"body"`
            URL    string `json:"url"`
        }{ OnCall: true, Title: title, Body: body, URL: "/alerts" }
        payload, _ := json.Marshal(note)
        // Send only to on-call users (skip if none)
        if len(tgtUserIDs) == 0 {
            log.Logf(req.Context(), "webpush: no on-call users; skipping send")
        } else {
            go sendPush(req.Context(), payload, tgtUserIDs...)
        }
    })
}
	mux.HandleFunc("POST /api/v2/site24x7/incoming", site24x7.Site24x7ToEventsAPI(app.AlertStore, app.IntegrationKeyStore))
	mux.HandleFunc("POST /api/v2/prometheusalertmanager/incoming", prometheus.PrometheusAlertmanagerEventsAPI(app.AlertStore, app.IntegrationKeyStore))

    // Wrap generic incoming to also fan out web push notifications
    {
        h := generic.ServeCreateAlert
        mux.HandleFunc("POST /api/v2/generic/incoming", func(w http.ResponseWriter, req *http.Request) {
            b, _ := io.ReadAll(io.LimitReader(req.Body, 1<<20))
            req.Body = io.NopCloser(bytes.NewReader(b))
            h(w, req)

            // Attempt to extract a title/body for the notification
            var p struct{
                Summary string `json:"summary"`
                Details string `json:"details"`
            }
            _ = json.Unmarshal(b, &p)
            if p.Summary == "" && p.Details == "" { return }
            svcID := permission.ServiceID(req.Context())
            var tgtUserIDs []string
            if svcID != "" {
                permission.SudoContext(req.Context(), func(ctx context.Context) {
                    if oc, err := app.OnCallStore.OnCallUsersByService(ctx, svcID); err == nil {
                        for _, u := range oc { tgtUserIDs = append(tgtUserIDs, u.UserID) }
                    } else {
                        log.Logf(ctx, "webpush: generic oncall lookup failed: %v", err)
                    }
                })
            }
            log.Logf(req.Context(), "webpush: generic incoming; serviceID=%s title=%q body.len=%d target-users=%d", svcID, p.Summary, len(p.Details), len(tgtUserIDs))
            note := struct{
                Title string `json:"title"`
                Body  string `json:"body"`
                URL   string `json:"url"`
            }{ Title: p.Summary, Body: p.Details, URL: "/alerts" }
            payload, _ := json.Marshal(note)
            // Send only to on-call users (skip if none)
            if len(tgtUserIDs) == 0 {
                log.Logf(req.Context(), "webpush: no on-call users; skipping send")
            } else {
                go sendPush(req.Context(), payload, tgtUserIDs...)
            }
        })
    }
	mux.HandleFunc("POST /api/v2/heartbeat/{heartbeatID}", generic.ServeHeartbeatCheck)
	mux.HandleFunc("GET /api/v2/user-avatar/{userID}", generic.ServeUserAvatar)
	mux.HandleFunc("GET /api/v2/calendar", app.CalSubStore.ServeICalData)

	mux.HandleFunc("POST /api/v2/twilio/message", app.twilioSMS.ServeMessage)
	mux.HandleFunc("POST /api/v2/twilio/message/status", app.twilioSMS.ServeStatusCallback)
	mux.HandleFunc("POST /api/v2/twilio/call", app.twilioVoice.ServeCall)
	mux.HandleFunc("POST /api/v2/twilio/call/status", app.twilioVoice.ServeStatusCallback)

	mux.HandleFunc("POST /api/v2/slack/message-action", app.slackChan.ServeMessageAction)

	middleware = append(middleware,
		httpRewrite(app.cfg.HTTPPrefix, "/v1/graphql2", "/api/graphql"),
		httpRedirect(app.cfg.HTTPPrefix, "/v1/graphql2/explore", "/api/graphql/explore"),

		httpRewrite(app.cfg.HTTPPrefix, "/v1/config", "/api/v2/config"),
		httpRewrite(app.cfg.HTTPPrefix, "/v1/identity/providers", "/api/v2/identity/providers"),
		httpRewrite(app.cfg.HTTPPrefix, "/v1/identity/providers/", "/api/v2/identity/providers/"),
		httpRewrite(app.cfg.HTTPPrefix, "/v1/identity/logout", "/api/v2/identity/logout"),

		httpRewrite(app.cfg.HTTPPrefix, "/v1/webhooks/mailgun", "/api/v2/mailgun/incoming"),
		httpRewrite(app.cfg.HTTPPrefix, "/v1/webhooks/grafana", "/api/v2/grafana/incoming"),
		httpRewrite(app.cfg.HTTPPrefix, "/v1/api/alerts", "/api/v2/generic/incoming"),
		httpRewrite(app.cfg.HTTPPrefix, "/v1/api/heartbeat/", "/api/v2/heartbeat/"),
		httpRewriteWith(app.cfg.HTTPPrefix, "/v1/api/users/", func(req *http.Request) *http.Request {
			parts := strings.Split(strings.TrimSuffix(req.URL.Path, "/avatar"), "/")
			req.URL.Path = "/api/v2/user-avatar/" + parts[len(parts)-1]
			return req
		}),

		httpRewrite(app.cfg.HTTPPrefix, "/v1/twilio/sms/messages", "/api/v2/twilio/message"),
		httpRewrite(app.cfg.HTTPPrefix, "/v1/twilio/sms/status", "/api/v2/twilio/message/status"),
		httpRewrite(app.cfg.HTTPPrefix, "/v1/twilio/voice/call", "/api/v2/twilio/call?type=alert"),
		httpRewrite(app.cfg.HTTPPrefix, "/v1/twilio/voice/alert-status", "/api/v2/twilio/call?type=alert-status"),
		httpRewrite(app.cfg.HTTPPrefix, "/v1/twilio/voice/test", "/api/v2/twilio/call?type=test"),
		httpRewrite(app.cfg.HTTPPrefix, "/v1/twilio/voice/stop", "/api/v2/twilio/call?type=stop"),
		httpRewrite(app.cfg.HTTPPrefix, "/v1/twilio/voice/verify", "/api/v2/twilio/call?type=verify"),
		httpRewrite(app.cfg.HTTPPrefix, "/v1/twilio/voice/status", "/api/v2/twilio/call/status"),

		func(next http.Handler) http.Handler {
			twilioHandler := twilio.WrapValidation(
				// go back to the regular mux after validation
				twilio.WrapHeaderHack(next),
				*app.twilioConfig,
			)
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if strings.HasPrefix(req.URL.Path, "/api/v2/twilio/") {
					twilioHandler.ServeHTTP(w, req)
					return
				}

				next.ServeHTTP(w, req)
			})
		},
	)

	mux.HandleFunc("GET /health", app.healthCheck)
	mux.HandleFunc("GET /health/engine", app.engineStatus)
	mux.HandleFunc("GET /health/engine/cycle", app.engineCycle)
	mux.Handle("GET /health/", http.NotFoundHandler())

	webH, err := web.NewHandler(app.cfg.UIDir, app.cfg.HTTPPrefix)
	if err != nil {
		return err
	}

	// This is necessary so that we can return 404 for invalid/unknown API routes, otherwise it will get caught by the UI handler and incorrectly return the index.html or a 405 (Method Not Allowed) error.
	mux.Handle("GET /api/", http.NotFoundHandler())
	mux.Handle("POST /api/", http.NotFoundHandler())
	mux.Handle("GET /v1/", http.NotFoundHandler())
	mux.Handle("POST /v1/", http.NotFoundHandler())

	// non-API/404s go to UI handler and return index.html
	mux.Handle("GET /", webH)

	mux.Handle("GET /api/graphql/explore", webH)
	mux.Handle("GET /api/graphql/explore/", webH)

	mux.HandleFunc("GET /admin/riverui/", func(w http.ResponseWriter, r *http.Request) {
		err := permission.LimitCheckAny(r.Context(), permission.Admin)
		if permission.IsUnauthorized(err) {
			// render login since we're on a UI route
			webH.ServeHTTP(w, r)
			return
		}
		if errutil.HTTPError(r.Context(), w, err) {
			return
		}

		app.RiverUI.ServeHTTP(w, r)
	})
	mux.HandleFunc("POST /admin/riverui/api/", func(w http.ResponseWriter, r *http.Request) {
		err := permission.LimitCheckAny(r.Context(), permission.Admin)
		if errutil.HTTPError(r.Context(), w, err) {
			return
		}

		app.RiverUI.ServeHTTP(w, r)
	})

	app.srv = &http.Server{
		Handler: applyMiddleware(mux, middleware...),

		ReadHeaderTimeout: time.Second * 30,
		ReadTimeout:       time.Minute,
		WriteTimeout:      time.Minute,
		IdleTimeout:       time.Minute * 2,
		MaxHeaderBytes:    app.cfg.MaxReqHeaderBytes,
	}
	app.srv.Handler = promhttp.InstrumentHandlerInFlight(metricReqInFlight, app.srv.Handler)
	app.srv.Handler = promhttp.InstrumentHandlerCounter(metricReqTotal, app.srv.Handler)

	// Ingress/load balancer/proxy can do a keep-alive, backend doesn't need it.
	// It also makes zero downtime deploys nearly impossible; an idle connection
	// could have an in-flight request when the server closes it.
	app.srv.SetKeepAlivesEnabled(false)

	return nil
}
