package sqltrace

import (
	"context"
	"database/sql/driver"
	"time"

	"github.com/target/goalert/retry"
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
	var conn driver.Conn
	var err error
	err = retry.DoTemporaryError(func(_ int) error {
		conn, err = c.drv.Open(c.name)
		return err
	},
		retry.Log(ctx),
		retry.Context(ctx),
		retry.Limit(10),
		retry.FibBackoff(time.Second),
	)
	return conn, err
}
