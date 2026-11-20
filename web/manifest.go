package web

import (
	"encoding/json"
	"net/http"
	"strings"
)

type manifestIcon struct {
	Src     string `json:"src"`
	Sizes   string `json:"sizes"`
	Type    string `json:"type"`
	Purpose string `json:"purpose"`
}

type manifest struct {
	Name            string         `json:"name"`
	ShortName       string         `json:"short_name"`
	StartURL        string         `json:"start_url"`
	Scope           string         `json:"scope"`
	Display         string         `json:"display"`
	BackgroundColor string         `json:"background_color"`
	ThemeColor      string         `json:"theme_color"`
	Icons           []manifestIcon `json:"icons"`
}

// newManifest returns the web app manifest, which makes the UI installable as a PWA.
//
// The app opens on the alerts list, and all URLs honor the configured HTTP prefix.
func newManifest(appName, prefix string) manifest {
	prefix = strings.TrimSuffix(prefix, "/")

	return manifest{
		Name:            appName,
		ShortName:       appName,
		StartURL:        prefix + "/alerts",
		Scope:           prefix + "/",
		Display:         "standalone",
		BackgroundColor: "#ffffff",
		ThemeColor:      "#cc0000",
		Icons: []manifestIcon{
			{Src: prefix + "/static/favicon-192.png", Sizes: "192x192", Type: "image/png", Purpose: "any"},
			{Src: prefix + "/static/favicon-512.png", Sizes: "512x512", Type: "image/png", Purpose: "any"},
		},
	}
}

func serveManifest(w http.ResponseWriter, req *http.Request, appName, prefix string) {
	data, err := json.Marshal(newManifest(appName, prefix))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/manifest+json")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(data)
}
