package csp

import (
	"bytes"
	"mime"
	"net/http"
)

type nonceRW struct {
	http.ResponseWriter
	nonce string
}

func (w nonceRW) Write(b []byte) (int, error) {
	// check content type
	// if not html, return as-is
	ct := w.Header().Get("Content-Type")
	mediaType, _, _ := mime.ParseMediaType(ct) // ignore error, we just want the cleaned-up type
	if mediaType != "text/html" {
		return w.ResponseWriter.Write(b)
	}
	buf := make([]byte, len(b))
	copy(buf, b)
	buf = bytes.ReplaceAll(buf, []byte("<script"), []byte("<script nonce=\""+w.nonce+"\""))
	buf = bytes.ReplaceAll(buf, []byte("<style"), []byte("<style nonce=\""+w.nonce+"\""))
	buf = bytes.Replace(buf, []byte("<head>"), []byte(`<head><meta property="csp-nonce" content="`+w.nonce+`" />`), 1)
	_, err := w.ResponseWriter.Write(buf)
	return len(b), err
}

// NonceResponseWriter will add a nonce value to <script> and <style> tags written to the response.
func NonceResponseWriter(nonce string, w http.ResponseWriter) http.ResponseWriter {
	if nonce == "" {
		return w
	}

	return &nonceRW{ResponseWriter: w, nonce: nonce}
}
