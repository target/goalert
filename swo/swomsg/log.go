package swomsg

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/target/goalert/swo/swodb"
	"github.com/target/goalert/util/log"
)

// pollInterval is how often the log will be polled for new events.
const pollInterval = time.Second / 3

// Log is a reader for the switchover log.
type Log struct {
	pool *pgxpool.Pool

	lastLoad time.Time

	eventCh chan Message
}

// NewLog will create a new log reader, skipping any existing events.
func NewLog(ctx context.Context, pool *pgxpool.Pool) (*Log, error) {
	lastID, err := swodb.New(pool).LastLogID(ctx)
	if err != nil {
		return nil, err
	}

	l := &Log{
		pool:    pool,
		eventCh: make(chan Message),
	}
	go l.readLoop(log.FromContext(ctx).BackgroundContext(), lastID)
	return l, nil
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
			err = json.Unmarshal(e.Data, &w)
			if err != nil {
				log.Log(ctx, fmt.Errorf("error parsing event: %v", err))
				continue
			}
			w.TS = e.Timestamp.Time
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

	return swodb.New(l.pool).LogEvents(ctx, lastID)
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

	return l.pool.SendBatch(ctx, &b).Close()
}
