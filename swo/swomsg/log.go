package swomsg

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/util/sqlutil"
	"gorm.io/gorm"
)

type Log struct {
	db *gorm.DB
	id uuid.UUID

	readID int64

	events   chan []logEvent
	lastLoad time.Time
}

var ErrStaleLog = fmt.Errorf("cannot append until log is read")

type logEvent struct {
	ID        int64
	Timestamp time.Time
	Data      []byte
}

func NewLog(db *gorm.DB, id uuid.UUID) (*Log, error) {
	l := &Log{
		id:     id,
		db:     db.Table("switchover_log"),
		events: make(chan []logEvent, 1),
	}
	l.events <- nil

	return l, nil
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

func (l *Log) Next(ctx context.Context) (*Message, error) {
	events := <-l.events
	var err error
	for len(events) == 0 {
		events, err = l.loadEvents(ctx)
		if err != nil {
			l.events <- nil
			return nil, err
		}
	}

	var w Message
	err = json.Unmarshal(events[0].Data, &w)
	if err != nil {
		l.events <- events
		return nil, err
	}
	w.TS = events[0].Timestamp

	l.readID = events[0].ID
	l.events <- events[1:]

	return &w, nil
}

func (l *Log) loadEvents(ctx context.Context) ([]logEvent, error) {
	err := ctxSleep(ctx, time.Second-time.Since(l.lastLoad))
	if err != nil {
		return nil, err
	}
	l.lastLoad = time.Now()

	var events []logEvent
	err = l.db.
		Where("timestamp > now() - interval '1 minute'").
		Where("id > ?", l.readID).
		Order("id asc").
		Limit(100).
		Find(&events).Error
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (l *Log) Append(ctx context.Context, v interface{}) error {
	var msg Message
	switch m := v.(type) {
	case Ping:
		msg.Ping = &m
	case Ack:
		msg.Ack = &m
	case Reset:
		msg.Reset = &m
	case Error:
		msg.Error = &m
	case Execute:
		msg.Execute = &m
	case Plan:
		msg.Plan = &m
	case Progress:
		msg.Progress = &m
	case Done:
		msg.Done = &m
	case Hello:
		msg.Hello = &m
	default:
		return fmt.Errorf("unknown message type %T", m)
	}

	msg.ID = uuid.New()
	msg.NodeID = l.id
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	e := <-l.events
	err = l.db.WithContext(ctx).Exec("insert into switchover_log (id, timestamp, data) values ((select max(id)+1 from switchover_log), now(), ?)", data).Error
	l.events <- e

	if dbErr := sqlutil.MapError(err); dbErr != nil && dbErr.Code == "23505" {
		return ErrStaleLog
	}

	return err
}
