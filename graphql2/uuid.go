package graphql2

import (
	graphql "github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
)

func MarshalUUID(t uuid.UUID) graphql.Marshaler {
	if t == uuid.Nil {
		return graphql.Null
	}

	return graphql.MarshalString(t.String())
}

func UnmarshalUUID(v interface{}) (uuid.UUID, error) {
	str, err := graphql.UnmarshalString(v)
	if err != nil {
		return uuid.Nil, err
	}
	id, err := uuid.Parse(str)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
