package main

import (
	"embed"
	"fmt"
	"html/template"
)

//go:embed assets
var assets embed.FS

func buildForm(action, desc, id string) string {
	action = template.HTMLEscapeString(action)
	desc = template.HTMLEscapeString(desc)
	id = template.HTMLEscapeString(id)

	return fmt.Sprintf(`<form method="POST" class="link">
	<input type="hidden" name="action" value="%s">
	<input type="hidden" name="id" value="%s">
	<input type="submit" value="%s" title="%s">
</form>`, action, id, action, desc)
}

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{
	"action": func(action, desc, id string) template.HTML {
		return template.HTML(buildForm(action, desc, id))
	},
}).ParseFS(assets, "assets/*.html"))
