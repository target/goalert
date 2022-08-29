package mocktwilio

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type twError struct {
	Status  int    `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Info    string `json:"more_info"`
}

func respondErr(w http.ResponseWriter, err twError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)

	err.Info = "https://www.twilio.com/docs/errors/" + strconv.Itoa(err.Code)
	json.NewEncoder(w).Encode(err)
}

// IsStatusUpdateErr returns true if the error is from a status update.
func IsStatusUpdateErr(err error) bool {
	type statErr interface {
		IsStatusUpdate() bool
	}

	if err == nil {
		return false
	}

	e, ok := err.(statErr)
	return ok && e.IsStatusUpdate()
}
