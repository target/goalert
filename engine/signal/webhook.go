package signal

import (
	"encoding/json"
	"errors"
)

type UserWebhook struct {
	URL  string `json:"url,omitempty"`
	Body string `json:"body,omitempty"`
}

// Content returns a json marshalled UserWebhook object containing the signal message
func (wh UserWebhook) Content(m map[string]string) (res json.RawMessage, destVal string, message string, err error) {

	err = mapToStruct(m, &wh)
	if err != nil {
		return res, destVal, message, errors.New("UserWebhook mapToStruct error")
	}

	res, err = json.Marshal(wh)
	if err != nil {
		return res, destVal, message, errors.New("UserWebhook json marshal error")
	}

	return res, wh.URL, wh.Body, nil
}
