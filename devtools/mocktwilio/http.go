package mocktwilio

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

func (srv *Server) basePath() string {
	return "/2010-04-01/Accounts/" + srv.cfg.AccountSID
}

func (srv *Server) initHTTP() {
	srv.mux.HandleFunc(srv.basePath()+"/Messages.json", srv.HandleNewMessage)
	srv.mux.HandleFunc(srv.basePath()+"/Messages/", srv.HandleMessageStatus)
	srv.mux.HandleFunc(srv.basePath()+"/Calls.json", srv.HandleNewCall)
	// s.mux.HandleFunc(base+"/Calls/", s.serveCallStatus)
	// s.mux.HandleFunc("/v1/PhoneNumbers/", s.serveLookup)
}

func (s *Server) post(ctx context.Context, url string, v url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Twilio-Signature", Signature(s.cfg.AuthToken, url, v))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, errors.Errorf("non-2xx response: %s", resp.Status)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 && resp.StatusCode != 204 {
		return nil, errors.Errorf("non-204 response on empty body: %s", resp.Status)
	}

	return data, nil
}

// ServeHTTP implements the http.Handler interface for serving [mock] API requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if s.cfg.EnableAuth {
		user, pass, ok := req.BasicAuth()
		if !ok || user != s.cfg.AccountSID || pass != s.cfg.AuthToken {
			respondErr(w, twError{
				Status:  401,
				Code:    20003,
				Message: "Authenticate",
			})
			return
		}
	}

	s.mux.ServeHTTP(w, req)
}
