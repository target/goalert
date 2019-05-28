package mockslack

import (
	"encoding/json"
	"log"
	"net/http"
)

type response struct {
	OK   bool   `json:"ok"`
	Err  string `json:"error,omitempty"`
	Meta struct {
		Cursor string `json:"next_cursor,omitempty"`
	} `json:"response_metadata,omitempty"`
}

func respondWith(w http.ResponseWriter, data interface{}) {
	w.Header().Set("content-type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println("ERROR:", err)
	}
}

func respondErr(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}

	respondWith(w, err)
	return true
}

type middlewareFunc func(http.HandlerFunc) http.HandlerFunc

func middleware(h http.Handler, fns ...middlewareFunc) http.Handler {

	for i := range fns {
		// apply in reverse order
		h = fns[len(fns)-1-i](h.ServeHTTP)
	}

	return h
}
