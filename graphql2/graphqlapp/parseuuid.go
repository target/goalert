package graphqlapp

import (
	"github.com/google/uuid"
	"github.com/target/goalert/validation"
)

func parseUUID(fname, u string) (uuid.UUID, error) {
	id, err := uuid.Parse(u)
	if err != nil {
		return uuid.UUID{}, validation.NewFieldError(fname, "must be a valid UUID: "+err.Error())
	}

	return id, nil
}
