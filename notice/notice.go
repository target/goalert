package notice

//go:generate go tool stringer -type Type

import (
	"io"

	"github.com/99designs/gqlgen/graphql"
	"github.com/target/goalert/validation"
)

type Notice struct {
	Type    Type
	Message string
	Details string
}

// Type NoticeType represents the level of severity of a Notice.
type Type int

// Defaults to Warning when unset
const (
	TypeWarning Type = iota
	TypeError
	TypeInfo
)

// UnmarshalGQL implements the graphql.Marshaler interface
func (t *Type) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err != nil {
		return err
	}

	switch str {
	case "WARNING":
		*t = TypeWarning
	case "ERROR":
		*t = TypeError
	case "INFO":
		*t = TypeInfo
	default:
		return validation.NewFieldError("Type", "unknown type "+str)
	}

	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (t Type) MarshalGQL(w io.Writer) {
	switch t {
	case TypeWarning:
		graphql.MarshalString("WARNING").MarshalGQL(w)
	case TypeError:
		graphql.MarshalString("ERROR").MarshalGQL(w)
	case TypeInfo:
		graphql.MarshalString("INFO").MarshalGQL(w)
	}
}
