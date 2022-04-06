package main

import (
	"net/http"
	"sync"
)

type RR struct {
	h  []http.Handler
	n  int
	mx sync.Mutex
}

func (r *RR) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mx.Lock()
	handler := r.h[r.n]
	r.n = (r.n + 1) % len(r.h)
	r.mx.Unlock()
	handler.ServeHTTP(w, req)
}
