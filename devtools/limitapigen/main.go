package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/target/goalert/limit"
)

type SystemLimitInfo struct {
	ID          limit.ID
	Description string
	Value       int
}

type SystemLimitScanner struct {
	data []SystemLimitInfo
}

func (s *SystemLimitScanner) Visit(node ast.Node) ast.Visitor {
	v, ok := node.(*ast.ValueSpec)
	if !ok {
		return s
	}

	name, ok := v.Type.(*ast.Ident)
	if !ok || name.Name != "ID" {
		return s
	}

	var comments []string
	if v.Doc != nil {
		comments = make([]string, len(v.Doc.List))
		for index, val := range v.Doc.List {
			comments[index] = strings.TrimSpace(strings.TrimPrefix(val.Text, "//"))
		}
	}

	s.data = append(s.data, SystemLimitInfo{
		ID:          limit.ID(v.Names[0].Name),
		Description: strings.Join(comments, "\n"),
	})

	return s
}

var tmpl = template.Must(
	template.
		New("maplimit.go").
		Funcs(template.FuncMap{
			"quote": strconv.Quote,
		}).Parse(`
// Code generate by devtools/limitapigen DO NOT EDIT.

package graphql2
import (
	"github.com/target/goalert/limit"
)
// MapLimitValues will map a Limit struct into a flat list of SystemLimit structs.
func MapLimitValues(l limit.Limits) []SystemLimit {
	return []SystemLimit {
		{{- range . }}
		{ID: "{{.ID}}", Description: {{quote .Description}}, Value: l[limit.{{.ID}}]},
		{{- end}}
	}
}
// ApplyLimitValues will apply a list of LimitValues to a Limit struct.
func ApplyLimitValues(l limit.Limits, vals []SystemLimitInput) (limit.Limits, error) {
	for _, v := range vals {
		switch v.ID {
			{{- range .}}
		case "{{.ID}}":
			l[limit.{{.ID}}] = v.Value
		{{- end}}
		default:
			return l, validation.NewFieldError("ID", fmt.Sprintf("unknown limit ID '%s'", v.ID))	
		}
	}
	return l, nil
}

`))

func main() {
	out := flag.String("out", "", "Output file.")
	flag.Parse()

	w := os.Stdout
	if *out != "" {
		fd, err := os.Create(*out)
		if err != nil {
			panic(err)
		}
		defer fd.Close()
		w = fd
	}

	fset := token.NewFileSet() // positions are relative to fset
	// Parse src but stop after processing the imports.
	f, err := parser.ParseFile(fset, "../limit/id.go", nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
		return
	}

	s := SystemLimitScanner{}
	ast.Walk(&s, f)
	sort.Slice(s.data, func(i, j int) bool { return s.data[i].ID < s.data[j].ID })
	tmpl.Execute(w, s.data)
}
