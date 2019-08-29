package sqlutil

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"
	"github.com/target/goalert/util/log"
)

type Listener struct {
	notifCh      chan *pgx.Notification
	shutdownFunc func()
	shutdownCh   chan struct{}
	startCh      chan struct{}
	pauseCh      chan struct{}
	pauseStartCh  ???
	pauseStopCh   ???

	// refer to Lifecycle manager 
	ctx          context.Context
	mx           sync.Mutex
	paused       bool
}

// NewListener will create and initialize a Listener which will automatically reconnect and listen to the provided channels.
func NewListener(ctx context.Context, db *sql.DB, channels ...string) (*Listener, error) {
	ctx, cancel := context.WithCancel(ctx)
	l := &Listener{
		notifCh:      make(chan *pgx.Notification, 32),
		shutdownFunc: cancel,
		shutdownCh:   make(chan struct{}),
		ctx:          ctx,
	}

	conn, err := newListenConn(db, channels)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(l.notifCh)
		defer close(l.shutdownCh)
		t := time.NewTicker(3 * time.Second)
		defer t.Stop()

		for {
			err = l.handleNotifications(db, conn)
			if err != nil {
				log.Log(ctx, err)
			}
			// keep reconnecting
			for {
				select {
				case <-l.pauseCh:
					if !l.waitStart() {
						return
					}
				default:
				}
				select {
				case <-l.pauseCh:
					if !l.waitStart() {
						return
					}
				case <-ctx.Done():
					return
				case <-t.C:
				}
				conn, err = newListenConn(db, channels)
				if err == nil {
					break
				}
				// failed to reconnect
				log.Log(ctx, err)
			}
		}
	}()

	return l, nil
}

// Start will enable reconnections and messages.
func (l *Listener) Start() {

}

// Stop will end all current connections and stop reconnecting.
func (l *Listener) Stop() {
	l.mx.Lock()
	close(l.pauseCh)
	l.paused = true
	l.mx.Unlock()

	l.waitStart()
}

// Close will shut down the listener and returns after all connections have been completed.
// It is not necessary to call Stop() before Close().
func (l *Listener) Close() error {
	l.shutdownFunc()
	<-l.shutdownCh
	return nil
}

func (l *Listener) reset() context.Context {
	return nil
}

func (l *Listener) handleNotifications(db *sql.DB, conn *pgx.Conn) error {
	defer stdlib.ReleaseConn(db, conn)
	defer conn.Close()

	ctx, cancel := context.WithCancel(l.ctx)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		n, err := conn.WaitForNotification(ctx)
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

// NotificationChannel returns the notification channel for this listener.
// Nil values will not be returned until the listener is closed.
func (l *Listener) NotificationChannel() <-chan *pgx.Notification {
	return l.notifCh
}

func (l *Listener) isPaused() bool {
	l.mx.Lock()
	defer l.mx.Unlock()
	return l.paused
}

func (l *Listener) isPausedCh() <-chan bool {
	return nil
}

func newListenConn(db *sql.DB, channels []string) (*pgx.Conn, error) {
	conn, err := stdlib.AcquireConn(db)
	if err != nil {
		return nil, err
	}

	for _, name := range channels {
		err = conn.Listen(name)
		if err != nil || l.isPaused() {
			conn.Close()
			stdlib.ReleaseConn(db, conn)
			return nil, err
		}
	}
	return conn, nil
}

func (l *Listener) waitStart() bool {
	select {}
}
