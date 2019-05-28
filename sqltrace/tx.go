package sqltrace

import (
	"context"
	"database/sql/driver"

	"go.opencensus.io/trace"
)

type _Tx struct {
	conn *_Conn
	tx   driver.Tx
	ctx  context.Context
}

func (tx *_Tx) Rollback() error {
	_, sp := trace.StartSpan(tx.ctx, "SQL.Tx.Rollback")
	err := errSpan(tx.tx.Rollback(), sp)
	sp.End()
	tx.conn.span.End()
	tx.conn.span = nil
	return err
}
func (tx *_Tx) Commit() error {
	_, sp := trace.StartSpan(tx.ctx, "SQL.Tx.Commit")
	err := errSpan(tx.tx.Commit(), sp)
	sp.End()
	tx.conn.span.End()
	tx.conn.span = nil
	return err
}
