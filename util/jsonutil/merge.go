package jsonutil

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

// Merge will recursively merge a src JSON document into the dst JSON document.
func Merge(dst, src []byte) ([]byte, error) {
	var dstM, srcM map[string]interface{}
	if len(dst) == 0 {
		dstM = make(map[string]interface{})
	} else {
		err := json.Unmarshal(dst, &dstM)
		if err != nil {
			return nil, err
		}
	}
	err := json.Unmarshal(src, &srcM)
	if err != nil {
		return nil, err
	}

	err = applyValues(dstM, srcM)
	if err != nil {
		return nil, err
	}

	return json.Marshal(dstM)
}

func applyValues(dst, src map[string]interface{}, prefix ...string) error {
	for key, val := range src {
		if valMap, ok := val.(map[string]interface{}); ok {
			switch d := dst[key].(type) {
			case nil:
				dst[key] = valMap
			case map[string]interface{}:
				if err := applyValues(d, valMap, append(prefix, key)...); err != nil {
					return err
				}
			default:
				return errors.Errorf("schema type mismatch: expected %s.%s in DB to be map[string]interface{} but was %T", strings.Join(prefix, "."), key, d)
			}
			continue
		}
		dst[key] = val
	}

	return nil
}
