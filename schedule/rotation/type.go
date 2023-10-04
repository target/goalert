package rotation

import (
	"database/sql/driver"
	"fmt"
	"io"

	"github.com/target/goalert/validation"

	"github.com/99designs/gqlgen/graphql"
)

type Type string

const (
	TypeMonthly Type = "monthly"
	TypeWeekly  Type = "weekly"
	TypeDaily   Type = "daily"
	TypeHourly  Type = "hourly"
)

// Scan handles reading a Role from the DB format
func (r *Type) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		*r = Type(t)
	case string:
		*r = Type(t)
	default:
		return fmt.Errorf("could not process unknown type for rotation type: %T", t)
	}

	return nil
}

// Value converts the Role to the DB representation
func (r Type) Value() (driver.Value, error) {
	switch r {
	case TypeMonthly, TypeWeekly, TypeDaily, TypeHourly:
		return string(r), nil
	default:
		return nil, fmt.Errorf("unknown rotation type specified '%s'", r)
	}
}

// UnmarshalGQL implements the graphql.Marshaler interface
func (t *Type) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err != nil {
		return err
	}
	switch str {
	case "monthly":
		*t = TypeMonthly
	case "weekly":
		*t = TypeWeekly
	case "daily":
		*t = TypeDaily
	case "hourly":
		*t = TypeHourly
	default:
		return validation.NewFieldError("Type", "unknown rotation type "+str)
	}

	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (t Type) MarshalGQL(w io.Writer) {
	switch t {
	case TypeMonthly:
		graphql.MarshalString("monthly").MarshalGQL(w)
	case TypeWeekly:
		graphql.MarshalString("weekly").MarshalGQL(w)
	case TypeHourly:
		graphql.MarshalString("hourly").MarshalGQL(w)
	case TypeDaily:
		graphql.MarshalString("daily").MarshalGQL(w)
	}
}
