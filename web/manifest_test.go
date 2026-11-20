package web

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestServeManifest(t *testing.T) {
	for _, tc := range []struct{ prefix, start, scope, icon string }{
		{"", "/alerts", "/", "/static/favicon-192.png"},
		{"/goalert", "/goalert/alerts", "/goalert/", "/goalert/static/favicon-192.png"},
		{"/goalert/", "/goalert/alerts", "/goalert/", "/goalert/static/favicon-192.png"},
	} {
		rec := httptest.NewRecorder()
		serveManifest(rec, httptest.NewRequest("GET", "/manifest.webmanifest", nil), "MyAlert", tc.prefix)

		if ct := rec.Header().Get("Content-Type"); ct != "application/manifest+json" {
			t.Errorf("prefix %q: content-type = %q", tc.prefix, ct)
		}

		var m manifest
		if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
			t.Fatalf("prefix %q: %v", tc.prefix, err)
		}
		if m.Name != "MyAlert" || m.StartURL != tc.start || m.Scope != tc.scope || m.Icons[0].Src != tc.icon {
			t.Errorf("prefix %q: unexpected manifest: %+v", tc.prefix, m)
		}
	}
}
