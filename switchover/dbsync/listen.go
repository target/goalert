package dbsync

import (
	"context"
	"fmt"
	"github.com/target/goalert/switchover"
	"time"

	"github.com/jackc/pgx/stdlib"
)

func (s *Sync) listen() {
	for {
		// ignoring errors (will reconnect)
		err := func() error {
			c, err := stdlib.AcquireConn(s.oldDB)
			if err != nil {
				return err
			}
			defer stdlib.ReleaseConn(s.oldDB, c)

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
