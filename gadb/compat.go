package gadb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// SQLDBTX is an interface can be used for *sql.DB and *sql.Tx.
type SQLDBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func Compat(stdlib SQLDBTX) DBTX {
	return &compatDB{db: stdlib}
}

// NewCompat will create a new Queries instance using the provided *sql.DB or *sql.Tx.
//
// This is a compatability function while migrating away from *sql.DB.
func NewCompat(db SQLDBTX) *Queries {
	return New(Compat(db))
}

type compatDB struct {
	db SQLDBTX
}

func (c *compatDB) Exec(ctx context.Context, q string, args ...interface{}) (pgconn.CommandTag, error) {
	res, err := c.db.ExecContext(ctx, q, args...)
	if err != nil {
		return pgconn.CommandTag{}, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		// no row count so just return an empty tag.
		return pgconn.CommandTag{}, nil
	}

	oid, err := res.LastInsertId()
	if err != nil {
		// no insert ID so just "say" it's an update so that the row count is correct.
		return pgconn.NewCommandTag(fmt.Sprintf("UPDATE %d", count)), nil
	}

	return pgconn.NewCommandTag(fmt.Sprintf("INSERT %d %d", oid, count)), nil
}

func (c *compatDB) QueryRow(ctx context.Context, q string, args ...interface{}) pgx.Row {
	return c.db.QueryRowContext(ctx, q, args...)
}

type compatRows struct{ *sql.Rows }

func (c *compatRows) Close()                        { _ = c.Rows.Close() }
func (c *compatRows) CommandTag() pgconn.CommandTag { panic("not implemented") }
func (c *compatRows) Conn() *pgx.Conn               { panic("not implemented") }
func (c *compatRows) FieldDescriptions() []pgconn.FieldDescription {
	panic("not implemented")
}
func (c *compatRows) RawValues() [][]byte    { panic("not implemented") }
func (c *compatRows) Values() ([]any, error) { panic("not implemented") }

func (c *compatDB) Query(ctx context.Context, q string, args ...interface{}) (pgx.Rows, error) {
	rows, err := c.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	return &compatRows{Rows: rows}, nil
}
