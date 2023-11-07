package signal

import (
	"encoding/json"
	"errors"
)

type SlackChannel struct {
	ChannelID string `json:"channel_id,omitempty"`
	Message   string `json:"message,omitempty"`
}

// Content returns a json marshalled SlackChannel object containing the signal message
func (sc SlackChannel) Content(m map[string]string) (res json.RawMessage, destVal string, message string, err error) {

	err = mapToStruct(m, &sc)
	if err != nil {
		return res, destVal, message, errors.New("SlackChannel mapToStruct error")
	}

	res, err = json.Marshal(sc)
	if err != nil {
		return res, destVal, message, errors.New("SlackChannel json marshal error")
	}

	return res, sc.ChannelID, sc.Message, nil
}
