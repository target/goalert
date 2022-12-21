package graphql2

import (
	"io"
	"log"
	"strings"
	"time"

	graphql "github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
	"github.com/target/goalert/validation"
)

func MarshalISOTimestamp(t time.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		if t.IsZero() {
			if _, err := io.WriteString(w, "null"); err != nil {
				log.Println("ERROR: Issue with MarshalISOTimestamp when using io.WriteString:", err)
			}
			return
		}
		if _, err := io.WriteString(w, `"`+t.UTC().Format(time.RFC3339Nano)+`"`); err != nil {
			log.Println("ERROR: Issue with MarshalISOTimestamp when using io.WriteString:", err)
		}
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
