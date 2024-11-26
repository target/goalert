package pgxdb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// WithSQLDB will execute the provided function with a new Queries instance using the provided *sql.DB.
// This is a compatability function while migrating away from *sql.DB.
func WithSQLDB(ctx context.Context, db *sql.DB, fn func(*Queries) error) error {
	c, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("get db connection: %w", err)
	}
	defer c.Close()

	return c.Raw(func(driverConn any) error {
		conn := driverConn.(*pgx.Conn)
		return fn(New(conn))
	})
}
