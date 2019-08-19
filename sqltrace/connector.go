package sqltrace

import (
	"context"
	"database/sql/driver"
	"time"

	"github.com/target/goalert/retry"
	"go.opencensus.io/trace"
)

type _Connector struct {
	dbc driver.Connector
	drv *_Driver

	attrs []trace.Attribute
}

func (c *_Connector) Connect(ctx context.Context) (driver.Conn, error) {
	var conn driver.Conn
	var err error
	err = retry.DoTemporaryError(func(_ int) error {
		conn, err = c.dbc.Connect(ctx)
		return err
	},
		retry.Log(ctx),
		retry.Context(ctx),
		retry.Limit(10),
		retry.FibBackoff(time.Second/2),
	)
	return &_Conn{conn: conn, drv: c.drv, attrs: c.attrs}, err
}
func (c *_Connector) Driver() driver.Driver {
	return c.drv
}
