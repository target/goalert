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

func httpRedirect(prefix, from, to string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Path != from {
				next.ServeHTTP(w, req)
				return
			}

			http.Redirect(w, req, prefix+to, http.StatusTemporaryRedirect)
		})
	}
}

func httpRewriteWith(prefix, from string, fn func(req *http.Request) *http.Request) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Path == from || (strings.HasSuffix(from, "/") && strings.HasPrefix(req.URL.Path, from)) {
				req = fn(req)
				req.URL.Path = prefix + req.URL.Path
			}

			next.ServeHTTP(w, req)
		})
	}
}

func httpRewrite(prefix, from, to string) func(http.Handler) http.Handler {
	u, err := url.Parse(to)
	if err != nil {
		panic(err)
	}
	uQ := u.Query()

	return httpRewriteWith(prefix, from, func(req *http.Request) *http.Request {
		origPath := req.URL.Path
		req.URL.Path = u.Path
		if strings.HasSuffix(from, "/") {
			req.URL.Path += strings.TrimPrefix(origPath, from)
		}
		q := req.URL.Query()
		for key := range uQ {
			q.Set(key, uQ.Get(key))
		}
		req.URL.RawQuery = q.Encode()
		return req
	})
}
