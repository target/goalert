package mocktwilio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// handleCallStatus handles GET requests to /2010-04-01/Accounts/<AccountSid>/Calls/<CallSid>.json
func (srv *Server) handleCallStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondErr(w, twError{
			Status:  405,
			Code:    20004,
			Message: "Method not allowed",
		})
		return
	}

	id := strings.TrimPrefix(r.URL.Path, srv.basePath()+"/Calls/")
	id = strings.TrimSuffix(id, ".json")

	db := <-srv.callStateDB
	s := db[id]
	srv.callStateDB <- db

	if s == nil {
		respondErr(w, twError{
			Status:  404,
			Message: fmt.Sprintf("The requested resource %s was not found", r.URL.String()),
			Code:    20404,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}
