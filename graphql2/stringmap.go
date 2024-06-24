package graphql2

import (
	"fmt"

	graphql "github.com/99designs/gqlgen/graphql"
	"github.com/target/goalert/validation"
)

// MapValueError is an error type for map values.
type MapValueError struct {
	Key string
	Err error
}

func (m MapValueError) Error() string {
	return fmt.Sprintf("field %s: %s", m.Key, m.Err)
}

func UnmarshalStringMap(v interface{}) (map[string]string, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, validation.NewGenericError("must be a map")
	}
	res := make(map[string]string, len(m))
	for k, v := range m {
		str, err := graphql.UnmarshalString(v)
		if err != nil {
			return nil, MapValueError{Key: k, Err: err}
		}
		res[k] = str
	}

	return res, nil
}

func MarshalStringMap(v map[string]string) graphql.Marshaler {
	return graphql.MarshalAny(v)
}
