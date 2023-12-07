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
			return fmt.Sprintf("lower(REPLACE(REPLACE(%s, '_', ' '), '-', ' ')) ~ :~%s", columnName, argName)
		},
		"contains": func(argName string, columnName string) string {
			// search for the term in the column
			//
			// - case insensitive
			// - allow for partial matches
			// - escape % and _ using `\` (backslash -- the default escape character)
			return fmt.Sprintf(`%s ilike '%%' || REPLACE(REPLACE(REPLACE(:%s, '\', '\\'), '%%', '\%%'), '_', '\_') || '%%'`, columnName, argName)
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

// splitSearchTerms will separate the words present in search and return a slice with them
func splitSearchTerms(search string) []string {
	search = strings.ToLower(search)
	var terms []string
	var cur string
	for _, r := range search {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur += string(r)
			continue
		}
		if cur == "" {
			continue
		}

		terms = append(terms, cur)
		cur = ""
	}

	if cur != "" {
		terms = append(terms, cur)
	}

	return terms
}

// orderedPrefixRxFromTerms returns a PSQL regular expression that will match
// a string if:
//
// - it includes words with the all the given prefixes
//
// - those words are in the same order as the prefixes
func orderedPrefixRxFromTerms(prefixes []string) pgtype.Text {
	rx := "\\m" + strings.Join(prefixes, ".*\\m")

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
			terms := splitSearchTerms(val)
			args = append(args, orderedPrefixRxFromTerms(terms))
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
