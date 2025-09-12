package main

import (
	"testing"

	. "github.com/dave/jennifer/jen"
)

// func MapConfigValues(cfg config.Config) []ConfigValue {
// 	return []ConfigValue{
// 		{{- range .ConfigFields }}
// 		{ID: {{quote .ID}}, Type: {{.Type}}, Description: {{quote .Desc}}, Value: {{.Value}}{{if .Password}}, Password: true{{end}}{{if .Dep}}, Deprecated: {{quote .Dep}}, Title: {{quote .Title}}, Section: {{quote .Section}}{{end}}},
// 		{{- end}}
// 	}
// }

func TestGenConfigValues(t *testing.T) {
	t.Fatalf("%#v", GenConfigHints())
}

func TestGenMap(t *testing.T) {
	f := NewFile("graphql2")

	f.Func().Id("MapConfigValues").
		Params(Id("cfg").Qual("github.com/target/goalert/config", "Config")).
		Index().Id("ConfigValue").
		Block(
			Return(Index().Id("ConfigValue").Line().ValuesFunc(func(g *Group) {
				g.Add(
					Values(Dict{
						Id("ID"):          Lit("foo"),
						Id("Type"):        Id("ConfigTypeString"),
						Id("Description"): Lit("desc"),
					}),
				)
			})),
		)
	t.Fatalf("%#v", f)
}
