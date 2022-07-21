package swo

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v4/stdlib"
	"github.com/target/goalert/version"
)

type mgrConnector struct {
	dbc driver.Connector
}

var _ driver.Connector = (*mgrConnector)(nil)

func newMgrConnector(dbc driver.Connector) *mgrConnector {
	return &mgrConnector{dbc: dbc}
}

func (drv *mgrConnector) Driver() driver.Driver { return nil }

func (drv *mgrConnector) Connect(ctx context.Context) (driver.Conn, error) {
	c, err := drv.dbc.Connect(ctx)
	if err != nil {
		return nil, err
	}

	conn := c.(*stdlib.Conn)
	str, err := conn.Conn().PgConn().EscapeString(fmt.Sprintf("GoAlert %s (SWO Manager)", version.GitVersion()))
	if err != nil {
		conn.Close()
		return nil, err
	}

	_, err = conn.ExecContext(ctx, fmt.Sprintf("set application_name = '%s'", str), nil)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return c, nil
}
