package web

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// NewHandler creates a new http.Handler that will serve UI files
// using bundled assets or by proxying to urlStr if set.
func NewHandler(urlStr, prefix string) (http.Handler, error) {
	mux := http.NewServeMux()

	var extraScripts []string
	if urlStr == "" {
		mux.Handle("/static/", newMemoryHandler())
	} else {
		u, err := url.Parse(urlStr)
		if err != nil {
			return nil, errors.Wrap(err, "parse url")
		}
		proxy := httputil.NewSingleHostReverseProxy(u)
		mux.Handle("/static/", proxy)
		mux.Handle("/build/", proxy)

		// dev mode
		extraScripts = []string{"../build/vendorPackages.dll.js"}
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
		w.Header().Set("Cache-Control", "private; max-age=31536000, stale-while-revalidate=600, stale-if-error=259200")
		w.Header().Set("ETag", indexETag)
		http.ServeContent(w, req, "/", time.Time{}, bytes.NewReader(buf.Bytes()))
	})

	return mux, nil
}

type memoryHandler struct {
	files map[string]File
}
type memoryFile struct {
	*bytes.Reader
	name string
	size int
}

func (m *memoryHandler) Open(file string) (http.File, error) {
	if f, ok := m.files["src/build"+file]; ok {
		return &memoryFile{Reader: bytes.NewReader(f.Data()), name: f.Name, size: len(f.Data())}, nil
	}

	return nil, os.ErrNotExist
}
func (m *memoryHandler) ETag(url string) string {
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	url = path.Clean(url)

	f, ok := m.files["src/build"+url]
	if !ok {
		return ""
	}

	return `"sha256-` + f.Hash256() + `"`
}

func (m *memoryFile) Close() error { return nil }
func (m *memoryFile) Readdir(int) ([]os.FileInfo, error) {
	return nil, errors.New("not a directory")
}
func (m *memoryFile) Stat() (os.FileInfo, error) {
	return m, nil
}
func (m *memoryFile) Name() string      { return path.Base(m.name) }
func (m *memoryFile) Size() int64       { return int64(m.size) }
func (m *memoryFile) Mode() os.FileMode { return 0644 }
func (m *memoryFile) ModTime() time.Time {
	if strings.Contains(m.name, "/static/") {
		return time.Time{}
	}

	return time.Now()
}
func (m *memoryFile) IsDir() bool      { return false }
func (m *memoryFile) Sys() interface{} { return nil }

func newMemoryHandler() http.Handler {
	m := &memoryHandler{
		files: make(map[string]File, len(Files)),
	}
	for _, f := range Files {
		m.files[f.Name] = f
	}
	fs := http.FileServer(m)
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		etag := m.ETag(req.URL.Path)
		if etag != "" {
			w.Header().Set("ETag", etag)
			w.Header().Set("Cache-Control", "public; max-age=31536000, stale-while-revalidate=600, stale-if-error=259200")
		}
		fs.ServeHTTP(w, req)
	})
}
