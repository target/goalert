package main

import (
	"bytes"
	"net/http"
)

func (s *State) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/ui", s.HandleIndex)
	mux.HandleFunc("/ui/messages/", s.HandleMessages)
	mux.Handle("/ui/assets/", http.StripPrefix("/ui/", http.FileServer(http.FS(assets))))
}

func render(w http.ResponseWriter, tmplName string, data interface{}) {
	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, tmplName, data)
	if err != nil {
		var data struct {
			Error error
			Data  string
		}
		data.Error = err
		data.Data = buf.String()
		w.WriteHeader(500)
		tmpl.ExecuteTemplate(w, "error.html", data)
		return
	}

	w.Write(buf.Bytes())
}

func hasError(w http.ResponseWriter, req *http.Request, err error) bool {
	if err == nil {
		return false
	}

	q := req.URL.Query()
	q.Set("error", err.Error())

	http.Redirect(w, req, req.URL.Path+"?"+q.Encode(), http.StatusFound)
	return true
}
