package swomsg

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Log struct {
	db *gorm.DB
	id uuid.UUID

	readID int64

	events   []logEvent
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
		id: id,
		db: db.Table("switchover_log"),
	}

	// only ever load new events
	err := db.Table("switchover_log").Select("coalesce(max(id), 0)").Take(&l.readID).Error

	return l, err
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
	var err error
	for len(l.events) == 0 {
		err = l.loadEvents(ctx)
		if err != nil {
			return nil, err
		}
	}

	var w Message
	err = json.Unmarshal(l.events[0].Data, &w)
	if err != nil {
		return nil, err
	}
	w.TS = l.events[0].Timestamp

	l.readID = l.events[0].ID
	l.events = l.events[1:]

	return &w, nil
}

func (l *Log) loadEvents(ctx context.Context) error {
	err := ctxSleep(ctx, time.Second-time.Since(l.lastLoad))
	if err != nil {
		return err
	}
	l.lastLoad = time.Now()

	var events []logEvent
	err = l.db.
		WithContext(ctx).
		Where("timestamp > now() - interval '1 minute'").
		Where("id > ?", l.readID).
		Order("id asc").
		Limit(100).
		Find(&events).Error
	if err != nil {
		return err
	}

	l.events = append(l.events, events...)

	return nil
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
	l.db.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		err := db.Exec("lock switchover_log in exclusive mode").Error
		if err != nil {
			return err
		}

		return db.Exec("insert into switchover_log (id, timestamp, data) values (coalesce((select max(id)+1 from switchover_log), 1), now(), ?)", data).Error
	})

	return err
}
