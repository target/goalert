package swosync

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/swo/swoinfo"
)

type LogicalReplicator struct {
	srcConn *pgx.Conn
	dstConn *pgx.Conn

	tables   []swoinfo.Table
	seqNames []string

	progFn func(ctx context.Context, format string, args ...interface{})

	dstRows RowSet
}

func NewLogicalReplicator() *LogicalReplicator {
	return &LogicalReplicator{
		dstRows: make(RowSet),
	}
}

func (l *LogicalReplicator) SetSourceDB(db *pgx.Conn)      { l.srcConn = db }
func (l *LogicalReplicator) SetDestinationDB(db *pgx.Conn) { l.dstConn = db }

func (l *LogicalReplicator) SetProgressFunc(fn func(ctx context.Context, format string, args ...interface{})) {
	l.progFn = fn
}

func (l *LogicalReplicator) printf(ctx context.Context, format string, args ...interface{}) {
	if l.progFn == nil {
		return
	}

	l.progFn(ctx, format, args...)
}
