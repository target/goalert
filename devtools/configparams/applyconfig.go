package main

import (
	"reflect"

	"github.com/dave/jennifer/jen"
)

func ApplyConfig[T any](objName, cfgName, parserName string, typ T) *jen.Statement {
	t := reflect.TypeOf(typ)

	return jen.BlockFunc(func(g *jen.Group) {
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

				switch field.Type.Kind() {
				case reflect.Bool:
					g.Add(jen.Id(cfgName + "." + section.Name + "." + field.Name).Op("=").Id(parserName).Dot("ParseBool").Call())
				}
			}
		}
	})
}
