package sqldrv

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v4/stdlib"
)

// NewDB is a convenience function for creating a *sql.DB from a DB URL and application_name.
func NewDB(urlStr, appName string) (*sql.DB, error) {
	c, err := NewConnector(urlStr, appName)
	if err != nil {
		return nil, err
	}
	return sql.OpenDB(c), nil
}

// NewConnector will create a new driver.Connector with retry enabled and the provided application_name.
func NewConnector(urlStr, appName string) (driver.Connector, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("parse db url: %w", err)
	}
	q := u.Query()
	q.Set("application_name", appName)
	q.Set("enable_seqscan", "off")
	u.RawQuery = q.Encode()
	urlStr = u.String()

	return NewRetryDriver(&stdlib.Driver{}, 10).OpenConnector(urlStr)
}
