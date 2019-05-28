package alertlog

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type rawJSON json.RawMessage

func (r *rawJSON) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		buf := make([]byte, len(t))
		copy(buf, t)
		*r = rawJSON(buf)
	case nil:

	default:
		return fmt.Errorf("could not process unknown type %T", t)
	}

	return nil
}
func (r rawJSON) Value() (driver.Value, error) {
	if len(r) == 0 {
		return nil, nil
	}
	return []byte(r), nil
}
