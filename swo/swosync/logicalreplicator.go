package swosync

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/target/goalert/swo/swoinfo"
)

// LogicalReplicator manages synchronizing the source database to the destination database.
type LogicalReplicator struct {
	srcConn *pgx.Conn
	dstConn *pgx.Conn

	tables   []swoinfo.Table
	seqNames []string

	progFn func(ctx context.Context, format string, args ...interface{})
}

// NewLogicalReplicator creates a new LogicalReplicator.
func NewLogicalReplicator() *LogicalReplicator {
	return &LogicalReplicator{}
}

// SetSourceDB sets the source database and must be called before Start.
func (l *LogicalReplicator) SetSourceDB(db *pgx.Conn) { l.srcConn = db }

// SetDestinationDB sets the destination database and must be called before Start.
func (l *LogicalReplicator) SetDestinationDB(db *pgx.Conn) { l.dstConn = db }

// SetProgressFunc sets the function to call when progress is made, such as the currently syncing table.
func (l *LogicalReplicator) SetProgressFunc(fn func(ctx context.Context, format string, args ...interface{})) {
	l.progFn = fn
}

func (l *LogicalReplicator) printf(ctx context.Context, format string, args ...interface{}) {
	if l.progFn == nil {
		return
	}

	l.progFn(ctx, format, args...)
}
