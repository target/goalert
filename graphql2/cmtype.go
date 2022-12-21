package graphql2

import (
	"io"
	"strings"

	graphql "github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
	"github.com/target/goalert/user/contactmethod"
)

func MarshalContactMethodType(t contactmethod.Type) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, `"`+string(t)+`"`)
	})
}
func UnmarshalContactMethodType(v interface{}) (contactmethod.Type, error) {
	str, ok := v.(string)
	if !ok {
		return "", errors.New("timestamps must be strings")
	}
	str = strings.Trim(str, `"`)

	return contactmethod.Type(str), nil
}
