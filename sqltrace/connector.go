package sqltrace

import (
	"context"
	"database/sql/driver"

	"go.opencensus.io/trace"
)

type _Connector struct {
	dbc driver.Connector
	drv *_Driver

	attrs []trace.Attribute
}

func (c *_Connector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.dbc.Connect(ctx)
	return &_Conn{conn: conn, drv: c.drv, attrs: c.attrs}, err
}
func (c *_Connector) Driver() driver.Driver {
	return c.drv
}
