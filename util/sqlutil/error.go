package sqlutil

import (
	"strconv"

	"github.com/jackc/pgx"
	"github.com/lib/pq"
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
	if pqErr, ok := errors.Cause(err).(*pq.Error); ok {
		pos, _ := strconv.Atoi(pqErr.Position)
		return &Error{
			err:            err,
			Code:           string(pqErr.Code),
			Message:        pqErr.Message,
			Detail:         pqErr.Detail,
			Hint:           pqErr.Hint,
			TableName:      pqErr.Table,
			ConstraintName: pqErr.Constraint,
			Where:          pqErr.Where,
			ColumnName:     pqErr.Column,
			Position:       pos,
		}
	}
	if pgxErr, ok := errors.Cause(err).(pgx.PgError); ok {
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
