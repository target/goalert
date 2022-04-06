package swomsg

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/target/goalert/util/log"
)

const PollInterval = time.Second / 3

type Log struct {
	db *sql.DB

	readID int64

	lastLoad time.Time

	eventCh chan Message
}

var ErrStaleLog = fmt.Errorf("cannot append until log is read")

type logEvent struct {
	ID        int64
	Timestamp time.Time
	Data      []byte
}

func NewLog(ctx context.Context, db *sql.DB) (*Log, error) {
	var lastID int64
	// only ever load new events
	err := db.QueryRowContext(ctx, "select coalesce(max(id), 0) from switchover_log").Scan(&lastID)
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

func (l *Log) loadEvents(ctx context.Context, lastID int64) ([]logEvent, error) {
	err := ctxSleep(ctx, PollInterval-time.Since(l.lastLoad))
	if err != nil {
		return nil, err
	}
	l.lastLoad = time.Now()

	rows, err := l.db.QueryContext(ctx, "select id, timestamp, data from switchover_log where id > $1 order by id asc limit 100", lastID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []logEvent
	var r logEvent
	for rows.Next() {
		err := rows.Scan(&r.ID, &r.Timestamp, &r.Data)
		if err != nil {
			return nil, err
		}
		events = append(events, r)
	}

	return events, nil
}

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
	defer stdlib.ReleaseConn(l.db, conn)

	err = conn.SendBatch(ctx, &b).Close()
	if err != nil {
		return err
	}

	return nil
}
