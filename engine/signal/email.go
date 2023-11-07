package signal

import (
	"encoding/json"
	"errors"
)

type Email struct {
	Address string `json:"address,omitempty"`
	Subject string `json:"subject,omitempty"`
	Body    string `json:"body,omitempty"`
}

// Content returns a json marshalled Email object containing the signal message
func (e Email) Content(m map[string]string) (res json.RawMessage, destVal string, message string, err error) {

	err = mapToStruct(m, &e)
	if err != nil {
		return res, destVal, message, errors.New("SlackChannel mapToStruct error")
	}

	res, err = json.Marshal(e)
	if err != nil {
		return res, destVal, message, errors.New("SlackChannel json marshal error")
	}

	return res, e.Address, e.Body, nil
}
