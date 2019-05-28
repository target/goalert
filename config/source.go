package config

import "net/http"

// A Source will provide a snapshot of a Config struct.
type Source interface {
	Config() Config
}

// Static implements a config.Source that never changes it's values.
type Static Config

// Config will return the current value of s.
func (s Static) Config() Config { return Config(s) }

// Handler will return a new http.Handler that provides config to all requests.
func Handler(next http.Handler, src Source) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(w, req.WithContext(src.Config().Context(req.Context())))
	})
}
