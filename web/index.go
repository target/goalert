package web

import (
	"bytes"
	"encoding/json"
	"html/template"
	"strings"
	"time"

	"github.com/target/goalert/version"
)

// AppVersion returns the version string from `app.js` (if available).
func AppVersion() string {
	const searchStr = "var GOALERT_VERSION="
	for _, f := range Files {
		if f.Name != "src/build/static/app.js" {
			continue
		}
		data := f.Data()
		idx := bytes.Index(data, []byte(searchStr))
		if idx == -1 {
			return "err: version not found"
		}
		data = data[idx+len(searchStr):]
		idx = bytes.Index(data, []byte(";"))
		if idx == -1 {
			return "err: version unreadable"
		}
		data = data[:idx]
		var versionStr string
		err := json.Unmarshal(data, &versionStr)
		if err != nil {
			// ignore failures
			return "err: " + err.Error()
		}

		return versionStr
	}
	return ""
}

type renderData struct {
	// Prefix is the URL prefix for the GoAlert application.
	Prefix string

	// ExtraScripts can be used to load additional javascript files
	// before `app.js`.
	ExtraScripts []string
}

func (r renderData) PathPrefix() string   { return strings.TrimSuffix(r.Prefix, "/") }
func (r renderData) BuildStamp() string   { return version.BuildDate().UTC().Format(time.RFC3339) }
func (r renderData) GitCommit() string    { return version.GitCommit() }
func (r renderData) GitVersion() string   { return version.GitVersion() }
func (r renderData) GitTreeState() string { return version.GitTreeState() }

var indexTmpl = template.Must(template.New("index.html").Parse(`<!DOCTYPE html>
<html class="no-js" lang="en">
  <head>
    <meta charset="utf-8" />
    <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1" />
    <meta http-equiv="x-goalert-version" content="{{.GitVersion}}" />
    <meta http-equiv="x-goalert-build-date" content="{{.BuildStamp}}" />
    <meta http-equiv="x-goalert-git-commit" content="{{.GitCommit}}" />
    <meta http-equiv="x-goalert-git-tree-state" content="{{.GitTreeState}}" />
    
    <title>GoAlert</title>
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <link rel="preconnect" href="https://gravatar.com" />
    <link
      rel="shortcut icon"
      type="image/png"
      sizes="16x16"
      href="{{.Prefix}}/static/favicon-16.png"
    />
    <link
      rel="shortcut icon"
      type="image/png"
      sizes="32x32"
      href="{{.Prefix}}/static/favicon-32.png"
    />
    <link
      rel="shortcut icon"
      type="image/png"
      sizes="64x64"
      href="{{.Prefix}}/static/favicon-64.png"
    />
    <link
      rel="apple-touch-icon"
      type="image/png"
      href="{{.Prefix}}/static/favicon-192.png"
    />
  </head>
  <body>
    <div id="app"></div>
  <div id="graceful-unmount"></div>
  <script>
    pathPrefix = {{.PathPrefix}};
  </script>
	{{- $prefix := .Prefix}}
    {{- range .ExtraScripts}}
    <script src="{{$prefix}}/static/{{.}}"></script>
	{{- end}}
	<script src="{{.Prefix}}/static/app.js"></script>
  </body>
</html>
`))
