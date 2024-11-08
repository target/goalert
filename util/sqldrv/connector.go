package sqldrv

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5/stdlib"
)

// NewDB is a convenience function for creating a *sql.DB from a DB URL and application_name.
func NewDB(urlStr string) (*sql.DB, error) {
	c, err := NewConnector(urlStr)
	if err != nil {
		return nil, err
	}
	return sql.OpenDB(c), nil
}

// AppURL will add the application_name parameter to the provided URL.
func AppURL(urlStr, appName string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("parse db url: %w", err)
	}
	q := u.Query()
	q.Set("application_name", appName)
	q.Set("enable_seqscan", "off")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// NewConnector will create a new driver.Connector with retry enabled and the provided application_name.
func NewConnector(urlStr string) (driver.Connector, error) {
	return NewRetryDriver(&stdlib.Driver{}, 10).OpenConnector(urlStr)
}
