package sqlutil

import (
	"context"
	"errors"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgxlisten"
	"github.com/target/goalert/util/log"
)

// Listener will listen for NOTIFY commands on a set of channels.
type Listener struct {
	l      *pgxlisten.Listener
	cancel context.CancelCauseFunc

	mx sync.Mutex
	c  *sync.Cond

	isStopped bool
	resume    chan struct{}
}

// NewListener will create and initialize a Listener which will automatically reconnect and listen to the provided channels when running.
func NewListener(p *pgxpool.Pool) *Listener {
	l := &Listener{
		l: &pgxlisten.Listener{
			Connect: func(ctx context.Context) (*pgx.Conn, error) {
				// We use the same connection config as the pool, but without the pool. This is because we can't reuse the same connection for both listening and querying.
				return pgx.ConnectConfig(ctx, p.Config().ConnConfig)
			},
			LogError: log.Log,
		},
		resume: make(chan struct{}),
	}
	l.c = sync.NewCond(&l.mx)

	return l
}

// Handle will register a handler for a specific channel.
func (l *Listener) Handle(channel string, fn func(context.Context, string) error) {
	l.l.Handle(channel, pgxlisten.HandlerFunc(func(ctx context.Context, n *pgconn.Notification, conn *pgx.Conn) error {
		return fn(ctx, n.Payload)
	}))
}

var (
	errPause = errors.New("pause")
	errStop  = errors.New("stop")
)

// Run will start the listener and begin listening for notifications.
func (l *Listener) Run(ctx context.Context) {
	defer close(l.resume)

	for {
		cancelCtx, cancel := context.WithCancelCause(ctx)

		l.mx.Lock()
		l.cancel = cancel
		l.isStopped = false
		l.c.Broadcast()
		l.mx.Unlock()

		err := l.l.Listen(cancelCtx)

		l.mx.Lock()
		l.isStopped = true
		l.c.Broadcast()
		l.mx.Unlock()

		if errors.Is(context.Cause(cancelCtx), errStop) {
			return
		}
		if errors.Is(context.Cause(cancelCtx), errPause) {
			select {
			case <-ctx.Done():
				return
			case l.resume <- struct{}{}:
			}
			continue
		}

		log.Log(ctx, err)
		return
	}
}

// Start will enable reconnections and messages.
func (l *Listener) Start() { <-l.resume }

// Stop will end all current connections and stop reconnecting.
func (l *Listener) Stop() {
	l.mx.Lock()
	defer l.mx.Unlock()

	l.cancel(errPause)

	for !l.isStopped {
		l.c.Wait()
	}
}

// Shutdown will shut down the listener and returns after all connections have been completed.
// It is not necessary to call Stop() before Close().
func (l *Listener) Shutdown(context.Context) error {
	l.mx.Lock()
	defer l.mx.Unlock()

	l.cancel(errStop)

	for !l.isStopped {
		l.c.Wait()
	}

	return nil
}
