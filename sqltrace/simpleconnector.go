package sqltrace

import (
	"context"
	"database/sql/driver"
)

type simpleConnector struct {
	name string
	drv  *_Driver
}

func newSimpleConnector(drv *_Driver, name string) (*simpleConnector, error) {
	return &simpleConnector{name: name, drv: drv}, nil
}
func (c *simpleConnector) Driver() driver.Driver {
	return c.drv
}
func (c *simpleConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return c.drv.Open(c.name)
}
