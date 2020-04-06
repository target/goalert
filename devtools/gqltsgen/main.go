package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile)
	var src []*ast.Source
	for _, file := range flag.Args() {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal("ERORR:", err)
		}
		src = append(src, &ast.Source{
			Name:  file,
			Input: string(data),
		})
	}

	doc, err := parser.ParseSchemas(src...)
	if err != nil {
		log.Fatal("ERORR:", err)
	}

	o := os.Stdout

	typeName := func(n string) string {
		switch n {
		case "String", "ID":
			return "string"
		case "Int":
			return "number"
		case "Boolean":
			return "boolean"
		}

		return n
	}

	fmt.Fprintf(o, "// Code generated with gqlparser DO NOT EDIT.\n\n")

	for _, def := range doc.Definitions {
		switch def.Kind {
		case ast.Enum:
			fmt.Fprintf(o, "enum %s {\n", def.Name)
			for _, e := range def.EnumValues {
				fmt.Fprintf(o, "\t%s = '%s',\n", e.Name, e.Name)
			}
			fmt.Fprintf(o, "}\n\n")
		case ast.InputObject, ast.Object:
			fmt.Fprintf(o, "export interface %s {\n", def.Name)
			for _, e := range def.Fields {
				mod := "?"
				if e.Type.NonNull {
					mod = ""
				}
				fmt.Fprintf(o, "\t%s: %s\n", e.Name+mod, typeName(e.Type.Name()))
			}
			fmt.Fprintf(o, "}\n\n")
		case ast.Scalar:
			fmt.Fprintf(o, "type %s = string\n\n", def.Name)
		default:
			log.Fatal("Unsupported kind:", def.Name, def.Kind)
		}
	}
}
