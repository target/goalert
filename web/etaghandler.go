package web

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/target/goalert/util/log"
)

type etagHandler struct {
	tags   map[string]string
	h      http.Handler
	fs     http.FileSystem
	mx     sync.Mutex
	static bool
}

func NewEtagFileServer(files http.FileSystem, static bool) http.Handler {
	return &etagHandler{
		tags:   make(map[string]string),
		h:      http.FileServer(files),
		fs:     files,
		static: static,
	}
}

func (e *etagHandler) etag(ctx context.Context, name string) string {
	e.mx.Lock()
	defer e.mx.Unlock()

	if tag, ok := e.tags[name]; e.static && ok {
		return tag
	}

	f, err := e.fs.Open(name)
	if err != nil {
		e.tags[name] = ""
		return ""
	}
	defer log.Close(ctx, f.Close)

	h := sha256.New()

	_, err = io.Copy(h, f)
	if err != nil {
		e.tags[name] = ""
		return ""
	}

	tag := fmt.Sprintf(`W/"sha256-%s"`, hex.EncodeToString(h.Sum(nil)))
	e.tags[name] = tag
	return tag
}

func (e *etagHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if tag := e.etag(req.Context(), req.URL.Path); tag != "" {
		if w.Header().Get("Cache-Control") == "" {
			w.Header().Set("Cache-Control", "public, max-age=60, stale-while-revalidate=600, stale-if-error=259200")
		}
		w.Header().Set("ETag", tag)
	}

	e.h.ServeHTTP(w, req)
}
