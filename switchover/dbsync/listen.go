package dbsync

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/target/goalert/switchover"
	"github.com/target/goalert/util/sqlutil"
)

func (s *Sync) listen(db *sql.DB) error {
	ctx := context.Background()
	l, err := sqlutil.NewListener(ctx, (*sqlutil.DBConnector)(db), switchover.StateChannel)
	if err != nil {
		return err
	}
	go func() {
		for n := range l.Notifications() {

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
	return nil

}
