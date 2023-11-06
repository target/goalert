package signal

import (
	"encoding/json"
	"fmt"
)

type UserWebhook struct {
	URL     string `json:"url,omitempty"`
	Message string `json:"message,omitempty"`
}

// fields returns a list of the UserWebhook struct's fields along with there corresponding json tag value
func (wh *UserWebhook) fields() ([]string, []interface{}) {
	return []string{"URL", "message"}, []interface{}{
		wh.URL,
		wh.Message,
	}
}

// Content returns a json marshalled UserWebhook object containing the signal message
func (wh UserWebhook) Content(m map[string]string) (res json.RawMessage, destVal string, message string, err error) {

	key, field := wh.fields()
	for i := 0; i < len(key); i++ {
		if val, ok := m[key[i]]; ok {
			field[i] = val
		} else {
			return res, destVal, message, fmt.Errorf("UserWebhook missing %s field", m[key[i]])
		}
	}

	return res, wh.URL, wh.Message, nil

	// res, err = json.Marshal(wh)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return res, destVal, message, errors.New("UserWebhook json marshal error")
	// }

	// return
}
