package swomsg

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/target/goalert/swo/swodb"
	"github.com/target/goalert/util/log"
)

// pollInterval is how often the log will be polled for new events.
const pollInterval = time.Second / 3

// Log is a reader for the switchover log.
type Log struct {
	db *sql.DB

	lastLoad time.Time

	eventCh chan Message
}

// NewLog will create a new log reader, skipping any existing events.
func NewLog(ctx context.Context, db *sql.DB) (*Log, error) {
	conn, err := stdlib.AcquireConn(db)
	if err != nil {
		return nil, err
	}
	defer releaseConn(db, conn)

	// only ever load new events
	lastID, err := swodb.New(conn).LastLogID(ctx)
	if err != nil {
		return nil, err
	}

	l := &Log{
		db:      db,
		eventCh: make(chan Message),
	}
	go l.readLoop(log.FromContext(ctx).BackgroundContext(), lastID)
	return l, nil
}

// releaseConn will release the current db conection
func releaseConn(db *sql.DB, conn *pgx.Conn) {
	_ = stdlib.ReleaseConn(db, conn)
}

// Events will return a channel that will receive all events in the log.
func (l *Log) Events() <-chan Message { return l.eventCh }

func (l *Log) readLoop(ctx context.Context, lastID int64) {
	for {
		events, err := l.loadEvents(ctx, lastID)
		if err != nil {
			log.Log(ctx, err)
			continue
		}

		for _, e := range events {
			lastID = e.ID
			var w Message
			err = json.Unmarshal(e.Data.Bytes, &w)
			if err != nil {
				log.Log(ctx, fmt.Errorf("error parsing event: %v", err))
				continue
			}
			w.TS = e.Timestamp
			l.eventCh <- w
		}
	}
}

func ctxSleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}

	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (l *Log) loadEvents(ctx context.Context, lastID int64) ([]swodb.SwitchoverLog, error) {
	err := ctxSleep(ctx, pollInterval-time.Since(l.lastLoad))
	if err != nil {
		return nil, err
	}
	l.lastLoad = time.Now()

	conn, err := stdlib.AcquireConn(l.db)
	if err != nil {
		return nil, err
	}
	defer releaseConn(l.db, conn)

	return swodb.New(conn).LogEvents(ctx, lastID)
}

// Append will append a message to the end of the log. Using an exclusive lock on the table, it ensures that each message will increment the log ID
// by exactly 1 with no gaps. All observers will see the messages in the same order.
func (l *Log) Append(ctx context.Context, msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	var b pgx.Batch
	b.Queue("begin")
	b.Queue("lock table switchover_log in exclusive mode")
	b.Queue("insert into switchover_log (id, timestamp, data) values (coalesce((select max(id)+1 from switchover_log), 1), now(), $1)", data)
	b.Queue("commit")
	b.Queue("rollback")

	conn, err := stdlib.AcquireConn(l.db)
	if err != nil {
		return err
	}
	defer releaseConn(l.db, conn)

	err = conn.SendBatch(ctx, &b).Close()
	if err != nil {
		return err
	}

	return nil
}
