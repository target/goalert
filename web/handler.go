package web

import (
	"bytes"
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
func NewHandler(urlStr string) (http.Handler, error) {
	if urlStr == "" {
		return newMemoryHandler(), nil
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "parse url")
	}

	return httputil.NewSingleHostReverseProxy(u), nil
}

type memoryHandler map[string]File
type memoryFile struct {
	*bytes.Reader
	file File
}

func (m memoryHandler) Open(file string) (http.File, error) {
	if f, ok := m["src/build"+file]; ok {
		return &memoryFile{Reader: bytes.NewReader(f.Data()), file: f}, nil
	}

	f, ok := m["src/build/index.html"]
	if !ok {
		return nil, errors.New("not found")
	}

	return &memoryFile{Reader: bytes.NewReader(f.Data()), file: f}, nil
}

func (m *memoryFile) Close() error { return nil }
func (m *memoryFile) Readdir(int) ([]os.FileInfo, error) {
	return nil, errors.New("not a directory")
}
func (m *memoryFile) Stat() (os.FileInfo, error) {
	return m, nil
}
func (m *memoryFile) Name() string      { return path.Base(m.file.Name) }
func (m *memoryFile) Size() int64       { return int64(len(m.file.Data())) }
func (m *memoryFile) Mode() os.FileMode { return 0644 }
func (m *memoryFile) ModTime() time.Time {
	if strings.Contains(m.file.Name, "/static/") {
		return time.Time{}
	}

	return time.Now()
}
func (m *memoryFile) IsDir() bool      { return false }
func (m *memoryFile) Sys() interface{} { return nil }

func rootFSFix(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			// necessary to avoid redirect loop
			req.URL.Path = "/alerts"
		}
		if strings.Contains(req.URL.Path, "/static/") {
			w.Header().Add("Cache-Control", "public, immutable, max-age=315360000")
		}

		h.ServeHTTP(w, req)
	})
}

func newMemoryHandler() http.Handler {
	if len(Files) > 0 {
		// preload
		go Files[0].Data()
	}
	m := make(memoryHandler, len(Files))
	for _, f := range Files {
		m[f.Name] = f
	}

	return rootFSFix(http.FileServer(m))
}
