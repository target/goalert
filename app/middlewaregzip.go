package app

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/felixge/httpsnoop"
)

var gzPool = sync.Pool{New: func() interface{} { return gzip.NewWriter(nil) }}

// wrapGzip will wrap an http.Handler to respond with gzip encoding.
func wrapGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") || req.Header.Get("Range") != "" {
			// Normal pass-through if gzip isn't accepted, there's no content type, or a Range is requested.
			//
			// Not going to handle the whole Transfer-Encoding vs Content-Encoding stuff -- just disable
			// gzip in this case.
			next.ServeHTTP(w, req)
			return
		}

		// If gzip is asked for, and we're not already replying with gzip
		// then wrap it. This is important as if we are proxying
		// UI assets (for example) we don't want to re-compress an already
		// compressed payload.

		var output io.Writer
		var check sync.Once
		cleanup := func() {}
		getOutput := func() {
			if w.Header().Get("Content-Encoding") != "" || w.Header().Get("Content-Type") == "" {
				// already encoded
				output = w
				return
			}

			gz := gzPool.Get().(*gzip.Writer)
			gz.Reset(w)
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Vary", "Accept-Encoding")
			w.Header().Del("Content-Length")
			cleanup = func() {
				_ = gz.Close()
				gzPool.Put(gz)
			}
			output = gz
		}

		ww := httpsnoop.Wrap(w, httpsnoop.Hooks{
			WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc { check.Do(getOutput); return next },
			Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
				return func(b []byte) (int, error) { check.Do(getOutput); return output.Write(b) }
			},
			ReadFrom: func(next httpsnoop.ReadFromFunc) httpsnoop.ReadFromFunc {
				return func(src io.Reader) (int64, error) { check.Do(getOutput); return io.Copy(output, src) }
			},
		})

		defer func() { cleanup() }()
		next.ServeHTTP(ww, req)
	})
}
