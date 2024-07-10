package graphql2

import (
	"io"

	graphql "github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func MarshalUUID(t uuid.UUID) graphql.Marshaler {
	if t == uuid.Nil {
		return graphql.Null
	}
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, t.String())
	})
}

func UnmarshalUUID(v interface{}) (uuid.UUID, error) {
	if str, ok := v.(string); ok {
		return uuid.Parse(str)
	}
	return uuid.Nil, errors.New("input must be an RFC-4122 formatted string")
}
