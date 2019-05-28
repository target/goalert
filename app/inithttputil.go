package app

import (
	"net/http"
	"net/url"
	"strings"
)

func applyMiddleware(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	// Needs to be wrapped in reverse order
	// so that the first one listed, is the "outermost"
	// handler, thus preserving the expected run-order.
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

func muxRedirect(mux *http.ServeMux, from, to string) {
	mux.HandleFunc(from, func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, to, http.StatusTemporaryRedirect)
	})
}
func muxRedirectPrefix(mux *http.ServeMux, prefix, to string) {
	mux.HandleFunc(prefix, func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, to+strings.TrimPrefix(req.URL.Path, prefix), http.StatusTemporaryRedirect)
	})
}
func muxRewriteWith(mux *http.ServeMux, from string, fn func(req *http.Request) *http.Request) {
	mux.HandleFunc(from,
		func(w http.ResponseWriter, req *http.Request) {
			mux.ServeHTTP(w, fn(req))
		})
}
func muxRewrite(mux *http.ServeMux, from, to string) {
	u, err := url.Parse(to)
	if err != nil {
		panic(err)
	}
	uQ := u.Query()

	muxRewriteWith(mux, from, func(req *http.Request) *http.Request {
		req.URL.Path = u.Path
		q := req.URL.Query()
		for key := range uQ {
			q.Set(key, uQ.Get(key))
		}
		req.URL.RawQuery = q.Encode()
		return req
	})
}
func muxRewritePrefix(mux *http.ServeMux, prefix, to string) {
	u, err := url.Parse(to)
	if err != nil {
		panic(err)
	}
	uQ := u.Query()
	muxRewriteWith(mux, prefix, func(req *http.Request) *http.Request {
		req.URL.Path = u.Path + strings.TrimPrefix(req.URL.Path, prefix)
		q := req.URL.Query()
		for key := range uQ {
			q.Set(key, uQ.Get(key))
		}
		req.URL.RawQuery = q.Encode()
		return req
	})
}
