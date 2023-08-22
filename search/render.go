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
	"unicode"

	"github.com/jackc/pgtype"
)

// RenderData is used as the data for a template with the ability to output a list
// of all possible arguments.
type RenderData interface {
	QueryArgs() []sql.NamedArg
}

// Helpers returns a map of all the helper functions that can be used in a template.
func Helpers() template.FuncMap {
	return template.FuncMap{
		"orderedPrefixSearch": func(argName string, columnName string) string {
			return fmt.Sprintf("lower(%s) ~ :~%s", columnName, argName)
		},
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

func orderedPrefixRxFromTerms(terms string) pgtype.Text {
	terms = strings.ToLower(terms)
	var rx string
	var cur string
	for _, r := range terms {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur += string(r)
			continue
		}
		if cur == "" {
			continue
		}

		// prefix match terms with the \m "word start" symbol
		if rx == "" {
			rx = "\\m" + cur
		} else {
			// extra words in between are allowed with .*
			rx = rx + ".*\\m" + cur
		}
		cur = ""
	}

	if cur != "" {
		if rx == "" {
			rx = "\\m" + cur
		} else {
			rx = rx + ".*\\m" + cur
		}
	}

	var t pgtype.Text
	_ = t.Set(rx)

	return t
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
	for i, arg := range nArgs {
		if strings.Contains(query, ":~"+arg.Name) {
			// regex match
			val, ok := arg.Value.(string)
			if !ok {
				return "", nil, fmt.Errorf("argument %d must be a string", i)
			}

			query = strings.ReplaceAll(query, ":~"+arg.Name, "$"+strconv.Itoa(n))
			args = append(args, orderedPrefixRxFromTerms(val))
			n++
		}
		rep := ":" + arg.Name
		if !strings.Contains(query, rep) {
			continue
		}
		query = strings.ReplaceAll(query, rep, "$"+strconv.Itoa(n))
		args = append(args, arg.Value)
		n++
	}
	return query, args, nil
}
