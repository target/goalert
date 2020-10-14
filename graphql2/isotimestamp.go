package graphql2

import (
	io "io"
	"strings"
	"time"

	graphql "github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
	"github.com/target/goalert/validation"
)

func MarshalISOTimestamp(t time.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		if t.IsZero() {
			io.WriteString(w, "null")
			return
		}
		io.WriteString(w, `"`+t.Format(time.RFC3339Nano)+`"`)
	})
}
func UnmarshalISOTimestamp(v interface{}) (time.Time, error) {
	str, ok := v.(string)
	if !ok {
		return time.Time{}, errors.New("timestamps must be strings")
	}
	str = strings.Trim(str, `"`)

	t, err := time.Parse(time.RFC3339Nano, str)
	return t, validation.WrapError(err)
}
