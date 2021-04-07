package web

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/pkg/errors"
)

//go:embed src/build
var bundleFS embed.FS

// NewHandler creates a new http.Handler that will serve UI files
// using bundled assets or by proxying to urlStr if set.
func NewHandler(urlStr, prefix string) (http.Handler, error) {
	mux := http.NewServeMux()

	etags := make(map[string]string)
	var mx sync.Mutex
	calcTag := func(name string, data []byte) string {
		mx.Lock()
		defer mx.Unlock()
		tag, ok := etags[name]
		if ok {
			return tag
		}
		sum := sha256.Sum256(data)
		tag = `W/"` + hex.EncodeToString(sum[:]) + `"`
		etags[name] = tag
		return tag
	}

	var extraScripts []string
	if urlStr == "" {
		mux.HandleFunc("/static/", func(w http.ResponseWriter, req *http.Request) {
			data, err := bundleFS.ReadFile(path.Join("src/build", req.URL.Path))
			if errors.Is(err, fs.ErrNotExist) {
				http.NotFound(w, req)
				return
			}

			w.Header().Set("Cache-Control", "public; max-age=60, stale-while-revalidate=600, stale-if-error=259200")
			w.Header().Set("ETag", calcTag(req.URL.Path, data))

			http.ServeContent(w, req, req.URL.Path, time.Time{}, bytes.NewReader(data))
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
