package sqlutil

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx"
	"github.com/target/goalert/switchover"
)

type ListenerEventType int

const (
	// ListenerEventConnected is emitted only when the database connection
	// has been initially initialized. The err argument of the callback
	// will always be nil.
	ListenerEventConnected ListenerEventType = iota

	// ListenerEventDisconnected is emitted after a database connection has
	// been lost, either because of an error or because Close has been
	// called. The err argument will be set to the reason the database
	// connection was lost.
	ListenerEventDisconnected

	// ListenerEventReconnected is emitted after a database connection has
	// been re-established after connection loss. The err argument of the
	// callback will always be nil. After this event has been emitted, a
	// nil pq.Notification is sent on the Listener.Notify channel.
	ListenerEventReconnected

	// ListenerEventConnectionAttemptFailed is emitted after a connection
	// to the database was attempted, but failed. The err argument will be
	// set to an error describing why the connection attempt did not
	// succeed.
	ListenerEventConnectionAttemptFailed
)

// Listener represents ....
type Listener struct {
	// Channel for receiving notifications from the database.  In some cases a
	// nil value will be sent.
	Notify chan *pgx.Notification

	EventCallback        EventCallbackType
	Name                 string
	MinReconnectInterval time.Duration
	maxReconnectInterval time.Duration
}

func NewListener(name string, minReconnectInterval time.Duration, maxReconnectInterval time.Duration, eventCallback EventCallbackType) *Listener {
	l := &Listener{
		eventCallback:        eventCallback,
		name:                 name,
		minReconnectInterval: minReconnectInterval,
		maxReconnectInterval: maxReconnectInterval,
		notify:               make(chan *pgx.Notification, 32),
	}

	return l
}

func (l *Listener) ListenerEventType(e EventCallbackType) EventCallbackType {
	return l.EventCallback
}

func (l *Listener) Listen(db *sql.DB) {
	for {
		// ignoring errors (will reconnect)
		err := func() error {
			c, err := getConfig()
			if err != nil {
				return err
			}

			c, err := pgx.ParseConnectionString(c.DBURL)
			if err != nil {
				return err
			}

			conn, err := pgx.Connect(c)
			if err != nil {
				return err
			}

			err = conn.Listen(switchover.StateChannel)
			if err != nil {
				return err
			}

			for {
				n, err := conn.WaitForNotification(context.Background())
				if err != nil {
					return err
				}
				stat, err := switchover.ParseStatus(n.Payload)
				if err != nil {
					fmt.Println("ERROR:", err)
					continue
				}

				s.mx.Lock()
				s.nodeStatus[stat.NodeID] = *stat
				s.mx.Unlock()
				select {
				case s.statChange <- struct{}{}:
				default:
				}
			}
		}()
		fmt.Println("ERROR:", err)
		time.Sleep(time.Second)
	}
}

func (l *Listener) Close() {
	// TODO Cleanup/Close on shutdown
}
