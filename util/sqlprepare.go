package util

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type ContextPreparer interface {
	PrepareContext(context.Context, string) (*sql.Stmt, error)
}

type PrepareStmt struct {
	*sql.Stmt
	q string
}

func (p *PrepareStmt) PrepareFor(ctx context.Context, cp ContextPreparer) (*sql.Stmt, error) {
	return cp.PrepareContext(ctx, p.q)
}

type Preparer interface {
	PrepareContext(context.Context, string) (*sql.Stmt, error)
}

// Prepare is used to prepare SQL statements.
//
// If Ctx is specified, it will be used to prepare all statements.
// Only the first error is recorded. Subsequent calls to `P` are
// ignored after a failure.
type Prepare struct {
	DB  Preparer
	Ctx context.Context
	Err error
}

type queryErr struct {
	q   string
	err *pq.Error
}
type QueryError interface {
	Query() string
	Cause() *pq.Error
}

func (q *queryErr) Query() string    { return q.q }
func (q *queryErr) Cause() *pq.Error { return q.err }
func (q *queryErr) Error() string    { return q.err.Error() }

func (p *Prepare) P(query string) (s *sql.Stmt) {
	if p.Err != nil {
		return nil
	}

	if p.Ctx != nil {
		s, p.Err = p.DB.PrepareContext(p.Ctx, query)
	} else {
		s, p.Err = p.DB.PrepareContext(context.Background(), query)
	}
	if p.Err != nil {
		if pqe, ok := p.Err.(*pq.Error); ok {
			p.Err = &queryErr{
				err: pqe,
				q:   query,
			}
		}

		p.Err = errors.WithStack(p.Err)
	}
	return s

}
