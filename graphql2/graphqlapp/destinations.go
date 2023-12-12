package graphqlapp

import (
	"context"

	"github.com/nyaruka/phonenumbers"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// TODO: sort destination types (i.e., disabled last)

func (q *Query) InputFieldValidate(ctx context.Context, dataType, value string) (bool, error) {
	switch dataType {
	case "PHONE":
		n, err := phonenumbers.Parse(value, "")
		if err != nil {
			return false, nil
		}
		return phonenumbers.IsValidNumber(n), nil
	case "EMAIL":
		return validate.Email("Email", value) == nil, nil
	case "URL":
		return validate.AbsoluteURL("URL", value) == nil, nil
	}

	return false, validation.NewGenericError("unsupported data type")
}

func (q *Query) DestinationType(ctx context.Context, typeID string) (*graphql2.DestinationTypeInfo, error) {
	types, err := q.DestinationTypes(ctx)
	if err != nil {
		return nil, err
	}

	for _, t := range types {
		if t.Type == typeID {
			return &t, nil
		}
	}

	return nil, nil
}
