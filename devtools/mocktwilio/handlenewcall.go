package mocktwilio

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ttacon/libphonenumber"
)

// HandleNewMessage handles POST requests to  /2010-04-01/Accounts/<AccountSid>/Calls.json
func (srv *Server) handleNewCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondErr(w, twError{
			Status:  405,
			Code:    20004,
			Message: "Method not allowed",
		})
		return
	}

	s := srv.newCallState()
	s.Direction = "outbound-api"
	s.To = r.FormValue("To")
	s.From = r.FormValue("From")
	s.StatusURL = r.FormValue("StatusCallback")
	s.CallURL = r.FormValue("Url")

	if s.CallURL == "" {
		respondErr(w, twError{
			Status:  400,
			Code:    21205,
			Message: "Url parameter is required.",
		})
		return
	}

	if s.To == "" {
		respondErr(w, twError{
			Status:  400,
			Code:    21201,
			Message: "No 'To' number is specified",
		})
		return
	}

	if s.From == "" {
		respondErr(w, twError{
			Status:  400,
			Code:    21213,
			Message: "No 'From' number is specified",
		})
		return
	}

	n := srv.number(s.From)
	if n == nil {
		respondErr(w, twError{
			Status:  400,
			Code:    21210,
			Message: fmt.Sprintf("The source phone number provided, %s, is not yet verified for your account. You may only make calls from phone numbers that you've verified or purchased from Twilio.", s.From),
		})
		return
	}

	if srv.number(s.To) != nil {
		// TODO: what's the correct approach here?
		http.Error(w, "app to app calls not implemented", http.StatusBadRequest)
		return
	}

	_, err := libphonenumber.Parse(s.To, "")
	if err != nil {
		respondErr(w, twError{
			Status:  400,
			Code:    13223,
			Message: fmt.Sprintf("The phone number you are attempting to call, %s, is not valid.", s.To),
		})
		return
	}

	if s.StatusURL != "" && !isValidURL(s.StatusURL) {
		respondErr(w, twError{
			Status:  400,
			Code:    21609,
			Message: fmt.Sprintf("The StatusCallback URL %s is not a valid URL.", s.StatusURL),
		})
		return
	}
	if !isValidURL(s.CallURL) {
		respondErr(w, twError{
			Status:  400,
			Code:    21205,
			Message: fmt.Sprintf("Url is not a valid URL: %s", s.CallURL),
		})
	}

	// Note: There is an inherent race condition where the first status update can be fired
	// before the original request returns, from the application's perspective, since there's
	// no way on the Twilio side to know if the application is ready for a status update.

	// marshal the return value before any status changes to ensure consistent return value
	data, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}

	db := <-srv.callStateDB
	db[s.ID] = s
	srv.callStateDB <- db

	srv.outboundCallCh <- s

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}
