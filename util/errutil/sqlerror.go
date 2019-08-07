package errutil

import (
	"github.com/jackc/pgx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// SQLError represents a driver-agnostic SQL error.
type SQLError struct {
	err            error
	Code           string
	Message        string
	Detail         string
	Hint           string
	TableName      string
	ConstraintName string
}

func (e SQLError) Error() string { return e.err.Error() }

// NewSQLError will return a SQLError from the given err object or nil otherwise.
func NewSQLError(err error) *SQLError {
	if pqErr, ok := errors.Cause(err).(*pq.Error); ok {
		return &SQLError{
			err:            err,
			Code:           string(pqErr.Code),
			Message:        pqErr.Message,
			Detail:         pqErr.Message,
			Hint:           pqErr.Hint,
			TableName:      pqErr.Table,
			ConstraintName: pqErr.Constraint,
		}
	}
	if pgxErr, ok := errors.Cause(err).(pgx.PgError); ok {
		return &SQLError{
			err:            err,
			Code:           pgxErr.Code,
			Message:        pgxErr.Message,
			Detail:         pgxErr.Message,
			Hint:           pgxErr.Hint,
			TableName:      pgxErr.TableName,
			ConstraintName: pgxErr.ConstraintName,
		}
	}
	return nil
}
