package web

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/target/goalert/app/csp"
	"github.com/target/goalert/config"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/version"
)

//go:embed src/build
var bundleFS embed.FS

//go:embed live.js
var liveJS []byte

//go:embed manifest.webmanifest
var pwaManifest []byte

//go:embed service-worker.js
var serviceWorker []byte

// validateAppJS will return an error if the app.js file is not valid or missing.
func validateAppJS(fs fs.FS) error {
	if version.GitVersion() == "dev" {
		// skip validation in dev mode
		return nil
	}

	fd, err := fs.Open("static/app.js")
	if err != nil {
		return fmt.Errorf("unable to open bundled app.js and ui-dir is unset, was make invoked with BUNDLE=1? (%w)", err)
	}
	defer fd.Close()

	// read first 512 bytes
	data, err := io.ReadAll(io.LimitReader(fd, 512))
	if err != nil {
		return fmt.Errorf("unable to read bundled app.js (%w)", err)
	}

	s := string(data)
	if !strings.HasPrefix(s, "var GOALERT_VERSION=") {
		return fmt.Errorf("bundled app.js is invalid, expected prefix \"var GOALERT_VERSION=\", got %q", s)
	}

	s, _, _ = strings.Cut(s, "\n") // only check first line
	if !strings.HasSuffix(s, ";") {
		return fmt.Errorf("bundled app.js is invalid, expected suffix \";\", got %q", s)
	}

	s = strings.TrimPrefix(s, "var GOALERT_VERSION=")
	s = strings.TrimSuffix(s, ";")
	var vers string
	err = json.Unmarshal([]byte(s), &vers)
	if err != nil {
		return fmt.Errorf("bundled app.js is invalid, expected quoted string, got %q (%w)", s, err)
	}

	if vers != version.GitVersion() {
		return fmt.Errorf("bundled app.js is invalid, version mismatch, expected %q, got %q", version.GitVersion(), vers)
	}

	return nil
}

// NewHandler creates a new http.Handler that will serve UI files
// using bundled assets or locally if uiDir if set.
func NewHandler(uiDir, prefix string) (http.Handler, error) {
	mux := http.NewServeMux()

	// in-memory push subscription store
	pushSubs := newPushStore()

	var extraJS string
	if uiDir != "" {
		extraJS = "/static/live.js"
		mux.Handle("/static/", NoCache(NewEtagFileServer(http.Dir(uiDir), false)))
		mux.HandleFunc("/static/live.js", func(w http.ResponseWriter, req *http.Request) {
			http.ServeContent(w, req, "/static/live.js", time.Time{}, bytes.NewReader(liveJS))
		})
	} else {
		sub, err := fs.Sub(bundleFS, "src/build")
		if err != nil {
			return nil, err
		}

		err = validateAppJS(sub)
		if err != nil {
			return nil, err
		}

		mux.Handle("/static/", NewEtagFileServer(http.FS(sub), true))
	}

	mux.HandleFunc("/manifest.webmanifest", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/manifest+json")
		http.ServeContent(w, req, "/manifest.webmanifest", time.Time{}, bytes.NewReader(pwaManifest))
	})
	mux.HandleFunc("/service-worker.js", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeContent(w, req, "/service-worker.js", time.Time{}, bytes.NewReader(serviceWorker))
	})

	// Minimal endpoint to accept and store push subscriptions.
	mux.HandleFunc("POST /api/push/subscribe", func(w http.ResponseWriter, req *http.Request) {
		var raw json.RawMessage
		data, err := io.ReadAll(io.LimitReader(req.Body, 1<<20)) // 1MB cap
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		raw = json.RawMessage(data)
		var tmp struct{
			Endpoint string `json:"endpoint"`
		}
		_ = json.Unmarshal(raw, &tmp)
		if tmp.Endpoint == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		pushSubs.add(pushSubscription{Endpoint: tmp.Endpoint, Raw: raw})
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("/api/graphql/explore", func(w http.ResponseWriter, req *http.Request) {
		cfg := config.FromContext(req.Context())

		serveTemplate(w, req, exploreTmpl, renderData{
			ApplicationName: cfg.ApplicationName(),
			Prefix:          prefix,
			ExtraJS:         extraJS,
			Nonce:           csp.NonceValue(req.Context()),
			VAPIDPublicKey:  cfg.WebPush.VAPIDPublicKey,
		})
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		cfg := config.FromContext(req.Context())

		serveTemplate(w, req, indexTmpl, renderData{
			ApplicationName: cfg.ApplicationName(),
			Prefix:          prefix,
			ExtraJS:         extraJS,
			Nonce:           csp.NonceValue(req.Context()),
			VAPIDPublicKey:  cfg.WebPush.VAPIDPublicKey,
		})
	})

	return mux, nil
}

func serveTemplate(w http.ResponseWriter, req *http.Request, tmpl *template.Template, data renderData) {
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if errutil.HTTPError(req.Context(), w, err) {
		return
	}

	nonceFree := make([]byte, buf.Len())
	copy(nonceFree, buf.Bytes())
	nonceFree = bytes.ReplaceAll(nonceFree, []byte(data.Nonce), nil)
	h := sha256.New()
	h.Write(nonceFree)
	etagValue := fmt.Sprintf(`W/"sha256-%s"`, hex.EncodeToString(h.Sum(nil)))
	w.Header().Set("ETag", etagValue)

	w.Header().Set("Cache-Control", "no-store")

	http.ServeContent(w, req, "/", time.Time{}, bytes.NewReader(buf.Bytes()))
}
