package sqltrace

import (
	"database/sql/driver"

	"go.opencensus.io/trace"
)

type _Rows struct {
	driver.Rows
	sp *trace.Span
}

func (r *_Rows) Next(dest []driver.Value) error {
	return errSpan(r.Rows.Next(dest), r.sp)
}
func (r *_Rows) Close() error {
	defer r.sp.End()
	return errSpan(r.Rows.Close(), r.sp)
}
