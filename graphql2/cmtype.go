package graphql2

import (
	"io"
	"log"
	"strings"

	"github.com/target/goalert/user/contactmethod"

	graphql "github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
)

func MarshalContactMethodType(t contactmethod.Type) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		if _, err := io.WriteString(w, `"`+string(t)+`"`); err != nil {
			log.Println("ERROR with MarshalContactMethodType when using io.WriteString:", err)
		}
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
