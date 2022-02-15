package graphqlapp

import (
	_ "embed"
	"html/template"
)

const graphiqlVersion = "1.5.19"
const reactVersion = "17.0.2"

//go:embed playground.html
var playHTML string

var playTmpl = template.Must(template.New("graphqlPlayground").Parse(playHTML))
