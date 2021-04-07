package web

import (
	"bytes"
	_ "embed"
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

//go:embed index.html
var indexHTML string

var indexTmpl = template.Must(template.New("index.html").Parse(indexHTML))
