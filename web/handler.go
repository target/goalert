package web

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

//go:embed src/build
var bundleFS embed.FS

// NewHandler creates a new http.Handler that will serve UI files
// using bundled assets or by proxying to urlStr if set.
func NewHandler(urlStr, prefix string) (http.Handler, error) {
	mux := http.NewServeMux()

	var extraScripts []string
	if urlStr == "" {
		fServ := http.FileServer(http.FS(bundleFS))
		mux.HandleFunc("/static/", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Cache-Control", "public; max-age=60, stale-while-revalidate=600, stale-if-error=259200")
			fServ.ServeHTTP(w, req)
		})
	} else {
		u, err := url.Parse(urlStr)
		if err != nil {
			return nil, errors.Wrap(err, "parse url")
		}
		proxy := httputil.NewSingleHostReverseProxy(u)
		mux.Handle("/static/", proxy)
		mux.Handle("/build/", proxy)

		// dev mode
		extraScripts = []string{"vendor.js"}
	}

	var buf bytes.Buffer
	err := indexTmpl.Execute(&buf, renderData{
		Prefix:       prefix,
		ExtraScripts: extraScripts,
	})
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	h.Write(buf.Bytes())
	indexETag := fmt.Sprintf(`"sha256-%s"`, hex.EncodeToString(h.Sum(nil)))

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Cache-Control", "private; max-age=60, stale-while-revalidate=600, stale-if-error=259200")
		w.Header().Set("ETag", indexETag)
		http.ServeContent(w, req, "/", time.Time{}, bytes.NewReader(buf.Bytes()))
	})

	return mux, nil
}
