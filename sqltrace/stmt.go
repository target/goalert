package sqltrace

import (
	"context"
	"database/sql/driver"
	"io"

	"github.com/target/goalert/util/sqlutil"
	"go.opencensus.io/trace"
)

type _Stmt struct {
	driver.Stmt
	query string
	conn  *_Conn
}

var _ driver.Stmt = &_Stmt{}
var _ driver.StmtExecContext = &_Stmt{}
var _ driver.StmtQueryContext = &_Stmt{}

func getValue(args []driver.NamedValue) []driver.Value {
	values := make([]driver.Value, len(args))
	for i, arg := range args {
		values[i] = arg.Value
	}
	return values
}
func errSpan(err error, sp *trace.Span) error {
	if err == nil {
		return nil
	}
	if err == io.EOF {
		return err
	}

	attrs := []trace.Attribute{trace.BoolAttribute("error", true)}

	if e := sqlutil.MapError(err); e != nil {
		attrs = append(attrs,
			trace.StringAttribute("pq.error.detail", e.Detail),
			trace.StringAttribute("pq.error.hint", e.Hint),
			trace.StringAttribute("pq.error.code", e.Code),
			trace.StringAttribute("pq.error.table", e.TableName),
			trace.StringAttribute("pq.error.constraint", e.ConstraintName),
			trace.StringAttribute("pq.error.where", e.Where),
			trace.StringAttribute("pq.error.column", e.ColumnName),
		)
	}
	sp.Annotate(attrs, err.Error())

	return err
}

func (s *_Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (res driver.Result, err error) {
	ctx, sp := s.conn.startSpan(ctx, "SQL.Stmt.Exec")
	defer sp.End()
	s.conn.annotateSpan(s.query, args, sp)

	if sec, ok := s.Stmt.(driver.StmtExecContext); ok {
		res, err = sec.ExecContext(ctx, args)
	} else {
		//lint:ignore SA1019 We have to fallback if the wrapped driver doesn't implement StmtExecContext.
		res, err = s.Stmt.Exec(getValue(args))
	}
	errSpan(err, sp)

	return res, err
}

func (s *_Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (rows driver.Rows, err error) {
	ctx, sp := s.conn.startSpan(ctx, "SQL.Stmt.Query")
	s.conn.annotateSpan(s.query, args, sp)

	if sqc, ok := s.Stmt.(driver.StmtQueryContext); ok {
		rows, err = sqc.QueryContext(ctx, args)
	} else {
		//lint:ignore SA1019 We have to fallback if the wrapped driver doesn't implement StmtQueryContext.
		rows, err = s.Stmt.Query(getValue(args))
	}
	errSpan(err, sp)
	if err != nil {
		sp.End()
		return nil, err
	}

	return &_Rows{Rows: rows, sp: sp}, nil
}
