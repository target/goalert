package sqlutil

import (
	"github.com/jackc/pgconn"
	"github.com/pkg/errors"
)

// Error represents a driver-agnostic SQL error.
type Error struct {
	err            error
	Code           string
	Message        string
	Detail         string
	Hint           string
	TableName      string
	ConstraintName string
	Where          string
	ColumnName     string
	Position       int
}

func (e Error) Error() string { return e.err.Error() }

// MapError will return a Error from the given err object or nil otherwise.
func MapError(err error) *Error {

	if pgxErr, ok := errors.Cause(err).(*pgconn.PgError); ok {
		return &Error{
			err:            err,
			Code:           pgxErr.Code,
			Message:        pgxErr.Message,
			Detail:         pgxErr.Detail,
			Hint:           pgxErr.Hint,
			Where:          pgxErr.Where,
			TableName:      pgxErr.TableName,
			ConstraintName: pgxErr.ConstraintName,
			ColumnName:     pgxErr.ColumnName,
			Position:       int(pgxErr.Position),
		}
	}
	return nil
}
