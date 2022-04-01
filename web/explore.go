package web

import (
	_ "embed"
	"html/template"
)

//go:embed explore.html
var exploreHTML string
var exploreTmpl = template.Must(template.New("explore.html").Parse(exploreHTML))
