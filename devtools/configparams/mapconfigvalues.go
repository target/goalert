package main

import (
	"reflect"

	"github.com/dave/jennifer/jen"
	"github.com/target/goalert/config"
)

func fieldTitle(f reflect.StructField) string {
	title := f.Tag.Get("title")
	if title != "" {
		return title
	}

	return f.Name
}

func typeName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "ConfigTypeString"
	case reflect.Int:
		return "ConfigTypeInteger"
	case reflect.Bool:
		return "ConfigTypeBoolean"
		// string slice
	case reflect.Slice:
		if t.Elem().Kind() == reflect.String {
			return "ConfigTypeStringList"
		}
		panic("unknown slice type: " + t.String())
	default:
		panic("unknown type: " + t.String())
	}
}

func GenConfigHints() *jen.Statement {
	return jen.Func().Id("MapConfigHints").Params(jen.Id("cfg").Qual("github.com/target/goalert/config", "Hints")).
		Index().Id("ConfigHint").
		Block(
			jen.Return(jen.Index().Id("ConfigHint").Values(
				GenStructToValues("cfg", false, false, config.Hints{}),
			)))
}

func GenMapConfigValues() *jen.Statement {
	return jen.Func().Id("MapConfigValues").Params(jen.Id("cfg").Qual("github.com/target/goalert/config", "Config")).
		Index().Id("ConfigValue").
		Block(
			jen.Return(jen.Index().Id("ConfigValue").Values(
				GenStructToValues("cfg", true, false, config.Config{}),
			)))
}

func GenMapPublicConfigValues() *jen.Statement {
	return jen.Func().Id("MapPublicConfigValues").Params(jen.Id("cfg").Qual("github.com/target/goalert/config", "Config")).
		Index().Id("ConfigValue").
		Block(
			jen.Return(jen.Index().Id("ConfigValue").Values(
				GenStructToValues("cfg", true, true, config.Config{}),
			)))
}

func GenStructToValues[T any](argName string, full, publicOnly bool, typ T) *jen.Statement {
	t := reflect.TypeOf(typ)

	return jen.ValuesFunc(func(g *jen.Group) {
		for i := 0; i < t.NumField(); i++ {
			section := t.Field(i)
			if section.Type.Kind() != reflect.Struct {
				continue
			}
			for j := 0; j < section.Type.NumField(); j++ {
				field := section.Type.Field(j)
				if field.PkgPath != "" {
					// skip unexported fields
					continue
				}
				if publicOnly && field.Tag.Get("public") != "true" {
					continue
				}

				vals := jen.Dict{
					jen.Id("ID"):    jen.Lit(section.Name + "." + field.Name),
					jen.Id("Value"): jen.Id(argName + "." + section.Name + "." + field.Name),
				}
				if full {
					vals[jen.Id("Type")] = jen.Id(typeName(field.Type))
					vals[jen.Id("Title")] = jen.Lit(fieldTitle(field))
					vals[jen.Id("Section")] = jen.Lit(fieldTitle(section))
				}
				if field.Tag.Get("deprecated") != "" {
					vals[jen.Id("Deprecated")] = jen.Lit(field.Tag.Get("deprecated"))
				}
				if field.Tag.Get("info") != "" {
					vals[jen.Id("Description")] = jen.Lit(field.Tag.Get("info"))
				}
				if field.Tag.Get("password") == "true" {
					vals[jen.Id("Password")] = jen.Lit(true)
				}

				g.Add(
					jen.Line().Values(vals),
				)
			}
		}
	})
}

// 	return jen.Func().Id(name).Params(jen.Id("cfg").Qual("github.com/target/goalert/config", "Config")).
// 		Index().Id("ConfigValue").
// 		Block(
// 			jen.Return(jen.Index().Id("ConfigValue").
// 				)
// }

// {ID: {{quote .ID}}, Type: {{.Type}}, Description: {{quote .Desc}}, Value: {{.Value}}{{if .Password}}, Password: true{{end}}{{if .Dep}}, Deprecated: {{quote .Dep}}, Title: {{quote .Title}}, Section: {{quote .Section}}{{end}}},
