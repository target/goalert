package validate

import (
	"database/sql"
	"strconv"

	"github.com/google/uuid"
	"github.com/target/goalert/validation"
)

// UUID will validate a UUID, returning a FieldError
// if invalid.
func UUID(fname, u string) error {
	_, err := ParseUUID(fname, u)
	return err
}

// ParseUUID will validate a UUID, returning a FieldError
// if invalid and the parsed UUID otherwise.
func ParseUUID(fname, u string) (uuid.UUID, error) {
	if len(u) != 36 {
		// Format check only required to ensure string IDs are valid when being passed to DB.
		//
		// We can remove this check once we switch to uuid.UUID in structs everywhere.
		return uuid.UUID{}, validation.NewFieldError(fname, "must be valid UUID: format must be xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
	}
	parsed, err := uuid.Parse(u)
	if err != nil {
		return uuid.UUID{}, validation.NewFieldError(fname, "must be a valid UUID: "+err.Error())
	}
	return parsed, nil
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
	_, err := ParseManyUUID(fname, ids, max)
	return err
}

// ParseManyUUID will validate a slice of strings, checking each
// with the UUID validator, and returning a slice of the parsed UUIDs
// if successful.
func ParseManyUUID(fname string, ids []string, max int) ([]uuid.UUID, error) {
	if max != -1 && len(ids) > max {
		return nil, validation.NewFieldError(fname, "must not have more than "+strconv.Itoa(max))
	}
	if len(ids) == 0 {
		return nil, nil
	}
	uuids := make([]uuid.UUID, len(ids))
	errs := make([]error, 0, len(ids))
	var err error
	for i, id := range ids {
		uuids[i], err = ParseUUID(fname+"["+strconv.Itoa(i)+"]", id)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return nil, Many(errs...)
	}
	return uuids, nil
}
