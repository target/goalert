package graphqlapp

import (
	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/validation"
)

func parseUUID(fname, u string) (uuid.UUID, error) {
	id, err := uuid.FromString(u)
	if err != nil {
		return uuid.UUID{}, validation.NewFieldError(fname, "must be a valid UUID: "+err.Error())
	}

	return id, nil
}
