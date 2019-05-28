package graphql2

import (
	"github.com/target/goalert/user/contactmethod"
	"io"
	"strings"

	graphql "github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
)

func MarshalContactMethodType(t contactmethod.Type) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, `"`+string(t)+`"`)
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
