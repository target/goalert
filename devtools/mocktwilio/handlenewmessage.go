package mocktwilio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// HandleNewMessage handles POST requests to /Accounts/<AccountSid>/Messages.json
func (srv *Server) HandleNewMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondErr(w, twError{
			Status:  405,
			Code:    20004,
			Message: "Method not allowed",
		})
		return
	}

	s := srv.newSMS()
	s.Direction = "outbound-api"
	s.To = r.FormValue("To")
	s.From = r.FormValue("From")
	s.Body = r.FormValue("Body")
	s.MsgSID = r.FormValue("MessagingServiceSid")
	s.StatusURL = r.FormValue("StatusCallback")

	if s.Body == "" {
		respondErr(w, twError{
			Status:  400,
			Code:    21602,
			Message: "Message body is required.",
		})
		return
	}

	if s.To == "" {
		respondErr(w, twError{
			Status:  400,
			Code:    21604,
			Message: "A 'To' phone number is required.",
		})
		return
	}

	if s.StatusURL != "" {
		u, err := url.Parse(s.StatusURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			respondErr(w, twError{
				Status:  400,
				Code:    21609,
				Message: fmt.Sprintf("The StatusCallback URL %s is not a valid URL.", s.StatusURL),
			})
			return
		}
	}

	if strings.HasPrefix(s.From, "MG") && s.MsgSID == "" {
		s.MsgSID = s.ID
	}

	if s.MsgSID != "" {
		if len(srv.numberSvc(s.MsgSID)) == 0 {
			respondErr(w, twError{
				Status:  404,
				Message: fmt.Sprintf("The requested resource %s was not found", r.URL.String()),
				Code:    20404,
			})
			return
		}

		// API says you can't use both from and msg SID, but actual
		// implementation allows it and uses msg SID if present.
		s.From = ""
		s.Status = "accepted"
	} else {
		if s.From == "" {
			respondErr(w, twError{
				Status:  400,
				Message: "A 'From' phone number is required.",
				Code:    21603,
			})
			return
		}

		if srv.number(s.From) == nil {
			respondErr(w, twError{
				Status:  400,
				Message: fmt.Sprintf("The From phone number %s is not a valid, SMS-capable inbound phone number or short code for your account.", s.From),
				Code:    21606,
			})
			return
		}
		s.Status = "queued"
	}

	// Note: There is an inherent race condition where the first status update can be fired
	// before the original request returns, from the application's perspective, since there's
	// no way on the Twilio side to know if the application is ready for a status update.

	// marshal the return value before any status changes to ensure consistent return value
	data, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}

	srv.outboundSMSCh <- s

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}
