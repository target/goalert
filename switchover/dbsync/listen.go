package dbsync

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/stdlib"
	"github.com/target/goalert/switchover"
)

func (s *Sync) listen(db *sql.DB) {
	for {
		// ignoring errors (will reconnect)
		err := func() error {
			c, err := stdlib.AcquireConn(db)
			if err != nil {
				return err
			}
			defer stdlib.ReleaseConn(db, c)

			err = c.Listen(switchover.StateChannel)
			if err != nil {
				return err
			}

			for {
				n, err := c.WaitForNotification(context.Background())
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
