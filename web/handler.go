package web

import (
	"bytes"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/util/log"
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
		px := httputil.NewSingleHostReverseProxy(u)
		p := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			log.Logf(req.Context(), "PATH: %s", req.URL.Path)
			px.ServeHTTP(w, req)
		})
		mux.Handle("/static/", p)
		mux.Handle("/build/", p)

		// dev mode
		extraScripts = []string{"../build/vendorPackages.dll.js"}
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		err := indexTmpl.Execute(w, renderData{
			Prefix:       prefix,
			ExtraScripts: extraScripts,
		})
		if err != nil {
			log.Log(req.Context(), err)
		}
	})

	return mux, nil
}

type memoryHandler struct {
	files     map[string]File
	loadIndex sync.Once
	indexData []byte
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
	m := &memoryHandler{files: make(map[string]File, len(Files))}
	for _, f := range Files {
		m.files[f.Name] = f
	}

	return http.FileServer(m)
}
