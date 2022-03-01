package sqldrv

import (
	"context"
	"database/sql/driver"
)

// RetryDriver will wrap a driver.Driver so that all new connections will be
// retried on temporary errors.
type RetryDriver struct {
	drv   driver.Driver
	limit int
}

var (
	_ driver.Driver        = (*RetryDriver)(nil)
	_ driver.DriverContext = (*RetryDriver)(nil)
)

// NewRetryDriver returns a new RetryDriver with the provided connection retry limit.
func NewRetryDriver(drv driver.Driver, retryLimit int) *RetryDriver {
	if retryLimit == 0 {
		retryLimit = 10
	}
	return &RetryDriver{drv: drv, limit: retryLimit}
}

func (rd *RetryDriver) Open(name string) (driver.Conn, error) {
	cn, err := rd.OpenConnector(name)
	if err != nil {
		return nil, err
	}

	return cn.Connect(context.Background())
}

func (rd *RetryDriver) OpenConnector(name string) (driver.Connector, error) {
	dbc, ok := rd.drv.(driver.DriverContext)
	if !ok {
		return &retryConnector{name: name, drv: rd}, nil
	}

	cn, err := dbc.OpenConnector(name)
	if err != nil {
		return nil, err
	}
	return &retryConnector{dbc: cn, drv: rd}, nil
}
