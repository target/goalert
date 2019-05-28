package validate

import (
	"database/sql"
	"github.com/target/goalert/validation"
	"strconv"

	uuid "github.com/satori/go.uuid"
)

// UUID will validate a UUID, returning a FieldError
// if invalid.
func UUID(fname, u string) error {
	_, err := uuid.FromString(u)
	if err != nil {
		return validation.NewFieldError(fname, "must be a valid UUID: "+err.Error())
	}
	return nil
}

// NullUUID will validate a UUID, unless Null. It returns a FieldError
// if invalid.
func NullUUID(fname string, u sql.NullString) error {
	if !u.Valid {
		return nil
	}
	return UUID(fname, u.String)
}

// ManyUUID will validate a slice of strings, checking each
// with the UUID validator.
func ManyUUID(fname string, ids []string, max int) error {
	if max != -1 && len(ids) > max {
		return validation.NewFieldError(fname, "must not have more than "+strconv.Itoa(max))
	}
	errs := make([]error, 0, len(ids))
	var err error
	for i, id := range ids {
		err = UUID(fname+"["+strconv.Itoa(i)+"]", id)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return Many(errs...)
}
