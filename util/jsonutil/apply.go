package jsonutil

import "encoding/json"

// Apply will recursively merge a src value into the dst JSON document.
func Apply(dst []byte, src interface{}) ([]byte, error) {
	srcDoc, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	return Merge(dst, srcDoc)
}
