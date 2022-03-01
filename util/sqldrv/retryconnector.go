package sqldrv

import (
	"context"
	"database/sql/driver"
	"time"

	"github.com/target/goalert/retry"
)

type retryConnector struct {
	dbc  driver.Connector
	name string
	drv  *RetryDriver
}

var _ driver.Connector = (*retryConnector)(nil)

func (rc *retryConnector) Connect(ctx context.Context) (driver.Conn, error) {
	var conn driver.Conn
	var err error
	err = retry.DoTemporaryError(func(_ int) error {
		if rc.dbc == nil {
			conn, err = rc.dbc.Connect(ctx)
		} else {
			conn, err = rc.drv.Open(rc.name)
		}
		return err
	},
		retry.Log(ctx),
		retry.Context(ctx),
		retry.Limit(rc.drv.limit),
		retry.FibBackoff(time.Second/2),
	)
	return conn, err
}
func (c *retryConnector) Driver() driver.Driver { return c.drv }
