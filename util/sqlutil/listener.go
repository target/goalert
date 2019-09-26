package sqlutil

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
)

// Connector will return a new *pgx.Conn.
type Connector interface {
	Connect(context.Context) (*pgx.Conn, error)
}

// Releaser is an optional interface with a Release method to be called
// after a connection is closed.
type Releaser interface {
	Release(*pgx.Conn) error
}

// DBConnector implements the Connector and Releaser interfaces for
// a *sql.DB.
type DBConnector sql.DB

// Connect will return a *pgx.Conn from the underlying *sql.DB pool.
func (db *DBConnector) Connect(context.Context) (*pgx.Conn, error) {
	return stdlib.AcquireConn((*sql.DB)(db))
}

// Release will release a *pgx.Conn to the underlying *sql.DB pool.
func (db *DBConnector) Release(conn *pgx.Conn) error {
	return stdlib.ReleaseConn((*sql.DB)(db), conn)
}

// ConfigConnector implements the Connector interface for a `pgx.ConnConfig`.
type ConfigConnector pgx.ConnConfig

// Connect will get a new connection using the underlying `pgx.ConnConfig`.
func (cfg ConfigConnector) Connect(ctx context.Context) (*pgx.Conn, error) {
	return pgx.ConnectConfig(ctx, (*pgx.ConnConfig)(&cfg))
}

var (
	_ = Connector(&DBConnector{})
	_ = Releaser(&DBConnector{})
	_ = Connector(ConfigConnector{})
)

// Listener will listen for NOTIFY commands on a set of channels.
type Listener struct {
	notifCh chan *pgconn.Notification

	ctx      context.Context
	db       Connector
	channels []string

	stopFn    func()
	stoppedCh chan struct{}
	errCh     chan error

	mx      sync.Mutex
	conn    *pgx.Conn
	running bool
}

// NewListener will create and initialize a Listener which will automatically reconnect and listen to the provided channels.
func NewListener(ctx context.Context, db Connector, channels ...string) (*Listener, error) {
	l := &Listener{
		notifCh:  make(chan *pgconn.Notification, 32),
		ctx:      ctx,
		channels: channels,
		db:       db,
		errCh:    make(chan error),
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
		if err == context.Canceled {
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

// Close will shut down the listener and returns after all connections have been completed.
// It is not necessary to call Stop() before Close().
func (l *Listener) Close() error {
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
		err := l.connect(ctx)
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

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		n, err := l.conn.WaitForNotification(ctx)
		if err != nil {
			return err
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
	l.conn.Close(l.ctx)
	if r, ok := l.db.(Releaser); ok {
		r.Release(l.conn)
	}
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
	conn, err := l.db.Connect(ctx)
	if err != nil {
		return err
	}
	l.conn = conn

	for _, name := range l.channels {
		select {
		case <-ctx.Done():
			l.disconnect()
			return ctx.Err()
		default:
		}

		_, err = conn.Exec(ctx, "listen "+QuoteID(name))
		if err != nil {
			l.disconnect()
			return err
		}
	}
	l.conn = conn
	return nil
}
