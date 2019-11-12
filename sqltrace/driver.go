package sqltrace

import (
	"database/sql/driver"

	"github.com/jackc/pgx/v4/stdlib"
)

type _Driver struct {
	drv          driver.Driver
	includeQuery bool
	includeArgs  bool

	// TODO: remove once pgx supports specifying `sql.LevelRepeatableRead`
	// https://github.com/jackc/pgx/pull/572
	pgxRRFix bool
}

// WrapOptions allow specifying additional information to include in the trace.
type WrapOptions struct {
	Query bool // include the SQL query
	Args  bool // include the arguments passed
}

// WrapDriver will wrap a database driver with tracing information.
func WrapDriver(drv driver.Driver, opts *WrapOptions) driver.DriverContext {
	if opts == nil {
		opts = &WrapOptions{}
	}

	_, pgxRRFix := drv.(*stdlib.Driver)
	return &_Driver{drv: drv, includeArgs: opts.Args, includeQuery: opts.Query, pgxRRFix: pgxRRFix}
}

func (d *_Driver) Open(name string) (driver.Conn, error) {
	attrs, err := getConnAttributes(name)
	if err != nil {
		return nil, err
	}
	c, err := d.drv.Open(name)
	return &_Conn{conn: c, drv: d, attrs: attrs}, err
}

func (d *_Driver) OpenConnector(name string) (driver.Connector, error) {
	attrs, err := getConnAttributes(name)
	if err != nil {
		return nil, err
	}
	if dc, ok := d.drv.(driver.DriverContext); ok {
		dbc, err := dc.OpenConnector(name)
		return &_Connector{dbc: dbc, drv: d, attrs: attrs}, err
	}
	return newSimpleConnector(d, name)
}

func (d *_Driver) Driver() driver.Driver {
	return d
}
