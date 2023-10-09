package sqlutil

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"github.com/target/goalert/util/log"
)

// Listener will listen for NOTIFY commands on a set of channels.
type Listener struct {
	notifCh chan *pgconn.Notification

	logger *log.Logger

	ctx      context.Context
	db       *sql.DB
	channels []string

	stopFn    func()
	stoppedCh chan struct{}
	errCh     chan error

	mx      sync.Mutex
	conn    *sql.Conn
	running bool
}

// NewListener will create and initialize a Listener which will automatically reconnect and listen to the provided channels.
func NewListener(ctx context.Context, logger *log.Logger, db *sql.DB, channels ...string) (*Listener, error) {
	l := &Listener{
		notifCh:  make(chan *pgconn.Notification, 32),
		ctx:      ctx,
		channels: channels,
		db:       db,
		errCh:    make(chan error),
		logger:   logger,
	}

	err := l.connect(ctx)
	if err != nil {
		return nil, err
	}

	l.Start()

	return l, nil
}

// Start will enable reconnections and messages.
func (l *Listener) Start() {
	l.mx.Lock()
	defer l.mx.Unlock()

	if l.running {
		return
	}

	ctx, cancel := context.WithCancel(l.ctx)
	l.stopFn = cancel
	l.stoppedCh = make(chan struct{}, 1)

	go l.run(ctx)

	l.running = true
}

func (l *Listener) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			l.stoppedCh <- struct{}{}
			return
		default:
		}
		err := l.handleNotifications(ctx)
		if errors.Is(err, context.Canceled) {
			err = nil
		}
		if err != nil {
			l.errCh <- err
		}
	}
}

// Stop will end all current connections and stop reconnecting.
func (l *Listener) Stop() {
	l.mx.Lock()
	defer l.mx.Unlock()

	if !l.running {
		return
	}

	l.stopFn()
	<-l.stoppedCh

	l.running = false
}

// Close performs a shutdown with a background context.
func (l *Listener) Close() error {
	return l.Shutdown(l.logger.BackgroundContext())
}

// Shutdown will shut down the listener and returns after all connections have been completed.
// It is not necessary to call Stop() before Close().
func (l *Listener) Shutdown(context.Context) error {
	l.Stop()
	close(l.notifCh)
	close(l.errCh)
	return nil
}

func (l *Listener) handleNotifications(ctx context.Context) error {
	defer l.disconnect()
	t := time.NewTicker(3 * time.Second)
	defer t.Stop()
	for {
		err := l.connect(l.ctx)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err != nil {
			l.errCh <- err
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-t.C:
				continue
			}
		}
		break
	}

	return l.conn.Raw(func(c interface{}) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			n, err := c.(*stdlib.Conn).Conn().WaitForNotification(ctx)
			if err != nil {
				if ctx.Err() == nil {
					return errors.Wrap(err, "wait for notifications")
				}
				return nil
			}

			if n == nil {
				continue
			}
			select {
			// Writing to channel
			case l.notifCh <- n:
			default:
			}
		}
	})
}

// Errors will return a channel that will be fed errors from this listener.
func (l *Listener) Errors() <-chan error { return l.errCh }

// Notifications returns the notification channel for this listener.
// Nil values will not be returned until the listener is closed.
func (l *Listener) Notifications() <-chan *pgconn.Notification { return l.notifCh }

func (l *Listener) disconnect() {
	if l.conn == nil {
		return
	}
	l.conn.Close()
	l.conn = nil
}

func (l *Listener) connect(ctx context.Context) error {
	if l.conn != nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	conn, err := l.db.Conn(ctx)
	if err != nil {
		return errors.Wrap(err, "get connection")
	}

	l.conn = conn

	for _, name := range l.channels {
		select {
		case <-ctx.Done():
			l.disconnect()
			return ctx.Err()
		default:
		}

		_, err = conn.ExecContext(ctx, "listen "+QuoteID(name))
		if err != nil {
			l.disconnect()
			return err
		}
	}

	return nil
}
