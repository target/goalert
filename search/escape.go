package search

import "strings"

// We need to escape any characters that have meaning for `ILIKE` in Postgres.
// https://www.postgresql.org/docs/8.3/static/functions-matching.html
var escapeRep = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// Escape will escape a search string for use with the Postgres `like` and `ilike` operators.
func Escape(s string) string {
	return escapeRep.Replace(s)
}
