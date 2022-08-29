package main

import (
	"embed"
	"html/template"
)

//go:embed assets
var assets embed.FS

var tmpl = template.Must(template.New("").ParseFS(assets, "assets/*.html"))
