package search

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

// RenderData is used as the data for a template with the ability to output a list
// of all possible arguments.
type RenderData interface {
	QueryArgs() []sql.NamedArg
}

// Helpers returns a map of all the helper functions that can be used in a template.
func Helpers() template.FuncMap {
	return template.FuncMap{
		"textSearch": func(argName string, columnNames ...string) string {
			var buf strings.Builder
			buf.WriteRune('(')
			for i, columnName := range columnNames {
				if i > 0 {
					buf.WriteString(" OR ")
				}
				buf.WriteString(fmt.Sprintf("to_tsvector('english', replace(lower(%s), '.', ' ')) @@ plainto_tsquery('english', replace(lower(:%s),'.',' '))", columnName, argName))
			}

			buf.WriteRune(')')
			return buf.String()
		},
	}
}

// RenderQuery will render a search query with the given template and data.
// Named args in the format `:name:` will be replaced with the appropriate numbered
// args (e.g. `$1`, `$2`)
func RenderQuery(ctx context.Context, tmpl *template.Template, data RenderData) (query string, args []interface{}, err error) {
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", nil, err
	}

	nArgs := data.QueryArgs()
	sort.Slice(nArgs, func(i, j int) bool { return len(nArgs[i].Name) > len(nArgs[j].Name) })

	args = make([]interface{}, 0, len(nArgs))
	query = buf.String()
	n := 1
	for _, arg := range nArgs {
		rep := ":" + arg.Name
		if !strings.Contains(query, rep) {
			continue
		}
		query = strings.Replace(query, rep, "$"+strconv.Itoa(n), -1)
		args = append(args, arg.Value)
		n++
	}
	return query, args, nil
}
