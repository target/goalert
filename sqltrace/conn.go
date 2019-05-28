package sqltrace

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strconv"

	"go.opencensus.io/trace"
)

type _Conn struct {
	conn driver.Conn
	drv  *_Driver

	span *trace.Span

	attrs []trace.Attribute
}

var _ driver.Conn = &_Conn{}
var _ driver.ConnBeginTx = &_Conn{}
var _ driver.ConnPrepareContext = &_Conn{}
var _ driver.ExecerContext = &_Conn{}
var _ driver.QueryerContext = &_Conn{}

func (c *_Conn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}
func (c *_Conn) PrepareContext(ctx context.Context, query string) (stmt driver.Stmt, err error) {
	ctx, sp := c.startSpan(ctx, "SQL.Prepare")
	defer sp.End()
	c.annotateSpan(query, nil, sp)

	if cp, ok := c.conn.(driver.ConnPrepareContext); ok {
		stmt, err = cp.PrepareContext(ctx, query)
	} else {
		stmt, err = c.conn.Prepare(query)
	}
	errSpan(err, sp)
	if err != nil {
		return nil, err
	}

	return &_Stmt{
		query: query,
		Stmt:  stmt,
		conn:  c,
	}, nil
}

func (c *_Conn) startSpan(ctx context.Context, name string) (context.Context, *trace.Span) {
	if c.span != nil {
		return trace.StartSpanWithRemoteParent(ctx, name, c.span.SpanContext())
	}

	return trace.StartSpan(ctx, name)
}

func (c *_Conn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *_Conn) BeginTx(ctx context.Context, opts driver.TxOptions) (tx driver.Tx, err error) {
	ctx, sp := c.startSpan(ctx, "SQL.Tx")
	sp.AddAttributes(
		trace.BoolAttribute("sql.tx.readOnly", opts.ReadOnly),
		trace.Int64Attribute("sql.tx.isolation", int64(opts.Isolation)),
	)

	if cx, ok := c.conn.(driver.ConnBeginTx); ok {
		tx, err = cx.BeginTx(ctx, opts)
	} else {
		//lint:ignore SA1019 We have to fallback if the wrapped driver doesn't implement ConnBeginTx.
		tx, err = c.conn.Begin()
	}
	errSpan(err, sp)
	if err != nil {
		sp.End()
		return nil, err
	}
	c.span = sp
	return &_Tx{conn: c, tx: tx, ctx: ctx}, nil
}
func (c *_Conn) Close() error {
	return c.conn.Close()
}

func (c *_Conn) annotateSpan(query string, args []driver.NamedValue, sp *trace.Span) {
	sp.AddAttributes(c.attrs...)
	if c.drv.includeQuery {
		sp.AddAttributes(
			trace.StringAttribute("sql.query", query),
		)
	}
	if c.drv.includeArgs && len(args) > 0 {
		for _, arg := range args {
			if arg.Name == "" {
				arg.Name = "$" + strconv.Itoa(arg.Ordinal)
			}

			sp.AddAttributes(
				trace.StringAttribute("sql.arg["+strconv.Quote(arg.Name)+"]", fmt.Sprintf("%v", arg.Value)),
			)
		}
	}
}

func (c *_Conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (res driver.Result, err error) {
	cec, cecOk := c.conn.(driver.ExecerContext)
	//lint:ignore SA1019 We have to fallback if the wrapped driver doesn't implement ExecerContext.
	ce, ceOk := c.conn.(driver.Execer)
	if !cecOk && !ceOk {
		return nil, driver.ErrSkip
	}

	ctx, sp := c.startSpan(ctx, "SQL.Exec")
	defer sp.End()
	c.annotateSpan(query, args, sp)

	if cecOk {
		res, err = cec.ExecContext(ctx, query, args)
	} else {
		res, err = ce.Exec(query, getValue(args))
	}
	errSpan(err, sp)

	return res, err
}

func (c *_Conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (rows driver.Rows, err error) {
	cqc, cqcOk := c.conn.(driver.QueryerContext)
	//lint:ignore SA1019 We have to fallback if the wrapped driver doesn't implement QueryerContext.
	cq, cqOk := c.conn.(driver.Queryer)
	if !cqcOk && !cqOk {
		return nil, driver.ErrSkip
	}

	ctx, sp := c.startSpan(ctx, "SQL.Query")
	c.annotateSpan(query, args, sp)
	if cqcOk {
		rows, err = cqc.QueryContext(ctx, query, args)
	} else {
		rows, err = cq.Query(query, getValue(args))
	}
	errSpan(err, sp)
	if err != nil {
		sp.End()
		return nil, err
	}

	return &_Rows{Rows: rows, sp: sp}, nil
}
