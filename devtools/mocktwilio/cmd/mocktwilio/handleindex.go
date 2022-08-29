package main

import (
	"fmt"
	"net/http"

	"github.com/target/goalert/devtools/mocktwilio"
)

func (s *State) HandleIndex(w http.ResponseWriter, req *http.Request) {
	switch req.FormValue("action") {
	case "":
		var data struct {
			*State
			Error string
		}
		data.State = s
		data.Error = req.FormValue("error")
		render(w, "index.html", data)
		return
	case "sendSMS":
		msg, err := s.srv.SendMessage(req.Context(), req.FormValue("from"), s.FromNumber, req.FormValue("body"))
		if hasError(w, req, err) {
			return
		}

		s.sendSMS <- sendSMS{From: msg.From(), Body: msg.Text()}
	case "Add Secondary":
		s.srv.UpdateConfig(func(cfg mocktwilio.Config) mocktwilio.Config {
			cfg.SecondaryAuthToken = mocktwilio.NewAuthToken()
			return cfg
		})
	case "Remove":
		s.srv.UpdateConfig(func(cfg mocktwilio.Config) mocktwilio.Config {
			cfg.SecondaryAuthToken = ""
			return cfg
		})
	case "Promote":
		s.srv.UpdateConfig(func(cfg mocktwilio.Config) mocktwilio.Config {
			cfg.PrimaryAuthToken = cfg.SecondaryAuthToken
			cfg.SecondaryAuthToken = ""
			return cfg
		})
	default:
		hasError(w, req, fmt.Errorf("unknown action: %s", req.FormValue("action")))
		return
	}

	http.Redirect(w, req, req.URL.Path, http.StatusFound)
}
