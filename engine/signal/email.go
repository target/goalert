package signal

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Email struct {
	Subject string `json:"subject,omitempty"`
	Body    string `json:"body,omitempty"`
}

// fields returns a list of the Email struct's fields along with there corresponding json tag value
func (e *Email) fields() ([]string, []interface{}) {
	return []string{"subject", "body"}, []interface{}{
		e.Subject,
		e.Body,
	}
}

// Content returns a json marshalled Email object containing the signal message
func (e Email) Content(m map[string]string) (res json.RawMessage, destVal string, message string, err error) {

	key, field := e.fields()
	for i := 0; i < len(key); i++ {
		if val, ok := m[key[i]]; ok {
			field[i] = val
		} else {
			return res, destVal, message, fmt.Errorf("Email missing %s field", m[key[i]])
		}
	}

	// return res, wh.URL, wh.Message, nil

	res, err = json.Marshal(e)
	if err != nil {
		fmt.Println(err)
		return res, destVal, message, errors.New("Email json marshal error")
	}

	return
}
