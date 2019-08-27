package sqlutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx"
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
	Notify   chan *pgx.Notification
	Channels map[string]struct{}

	EventCallback        ListenerEventType
	Name                 string
	MinReconnectInterval time.Duration
	MaxReconnectInterval time.Duration
}

// EventCallbackType is the event callback type. See also ListenerEventType
// constants' documentation.
type EventCallbackType func(event ListenerEventType, err error)

func NewListener(name string, minReconnectInterval time.Duration, maxReconnectInterval time.Duration, eventCallback ListenerEventType) *Listener {
	l := &Listener{
		EventCallback:        eventCallback,
		Name:                 name,
		MinReconnectInterval: minReconnectInterval,
		MaxReconnectInterval: maxReconnectInterval,
		Notify:               make(chan *pgx.Notification, 32),
		Channels:             make(map[string]struct{}),
	}

	return l
}

func (l *Listener) Listen(db *sql.DB) {
	for {
		// TODO Don't ignore errors, perform logic for reconnect
		err := func() error {
			dbURL := os.Getenv("DB_URL")
			if dbURL == "" {
				dbURL = "postgres://goalert@127.0.0.1:5432?sslmode=disable"
			}

			dbCfg, err := pgx.ParseConnectionString(dbURL)
			if err != nil {
				return err
			}

			conn, err := pgx.Connect(dbCfg)
			if err != nil {
				return err
			}

			// TODO
			err = conn.Listen("channel_name_here")
			if err != nil {
				return err
			}

			for {
				n, err := conn.WaitForNotification(context.Background())
				if err != nil {
					return err
				}
				/*stat, err := switchover.ParseStatus(n.Payload)
				if err != nil {
					fmt.Println("ERROR:", err)
					continue
				}*/

				// WaitForNotification in a loop feeding a channel should get us the behavior we want,
				// then we can reconnect as-needed and just keep passing messages to the same channel
				// Use parsed notification message somewhere
				select {
				// Writing to channel
				case l.Notify <- n:
				default:
				}
			}
		}()
		fmt.Println("ERROR:", err)
		time.Sleep(time.Second)

		// Call Close() here?
	}
}

func (l *Listener) Close() {
	// TODO Cleanup/Close on shutdown

}
