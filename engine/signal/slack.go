package signal

import (
	"encoding/json"
	"fmt"
)

type SlackChannel struct {
	Message string `json:"message,omitempty"`
}

// fields returns a list of the SlackChannel struct's fields along with there corresponding json tag value
func (sc *SlackChannel) fields() ([]string, []interface{}) {
	return []string{"message"}, []interface{}{
		sc.Message,
	}
}

// Content returns a json marshalled SlackChannel object containing the signal message
func (sc SlackChannel) Content(m map[string]string) (res json.RawMessage, destVal string, message string, err error) {

	key, field := sc.fields()
	for i := 0; i < len(key); i++ {
		if val, ok := m[key[i]]; ok {
			field[i] = val
		} else {
			return res, destVal, message, fmt.Errorf("SlackChannel missing %s field", m[key[i]])
		}
	}

	return res, destVal, sc.Message, nil

	// res, err = json.Marshal(sc)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return res, destVal, message, errors.New("SlackChannel json marshal error")
	// }

	// return
}
