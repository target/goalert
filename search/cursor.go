package search

import (
	"encoding/base64"
	"encoding/json"
	"github.com/target/goalert/validation"
)

// ParseCursor will parse the data held in cursor c into the passed state object.
func ParseCursor(c string, state interface{}) error {
	data, err := base64.URLEncoding.DecodeString(c)
	if err != nil {
		return validation.NewFieldError("Cursor", err.Error())
	}
	err = json.Unmarshal(data, state)
	if err != nil {
		return validation.NewFieldError("Cursor", err.Error())
	}
	return nil
}

// Cursor will return a cursor for the given state data.
func Cursor(state interface{}) (string, error) {
	data, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(data), nil
}
