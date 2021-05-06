package jsonutil

import "encoding/json"

// Apply can be used in place of `json.Marshal` but with the effect of
// merging into the original document.
func Apply(original []byte, src interface{}) ([]byte, error) {
	srcDoc, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	return Merge(original, srcDoc)
}
