package sqlutil

import (
	"github.com/jackc/pgx/v5/pgconn"
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
	var pgxErr *pgconn.PgError

	if errors.As(err, &pgxErr) {
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

	var e Error
	if errors.As(err, &e) {
		return &e
	}

	return nil
}
