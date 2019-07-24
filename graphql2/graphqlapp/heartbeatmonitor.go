package graphqlapp

import (
	context "context"
	"database/sql"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/heartbeat"
)

type HeartbeatMonitor App

//func (a *App) HeartbeatMonitor() graphql2.HeartbeatResolver { return (*HeartbeatMonitor)(a) }

func (q *Query) HeartbeatMonitor(ctx context.Context, id string) (*heartbeatMonitor.monitor, error) {
	results, err := q.FindMany(ctx, id)
}

func (m *Mutation) CreateHeartbeatMonitor(ctx context.Context, input graphql2.CreateHeartbeatMonitorInput) (hb *heartbeat.Monitor, err error) {
	hb = &heartbeat.Monitor{
		ServiceID:      input.ServiceID,
		Name:           input.Name,
		TimeoutMinutes: input.TimeoutMinutes,
	}
	err = withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		hb, err = m.HeartbeatStore.CreateTx(ctx, tx, hb)
		return err
	})
	return hb, err
}

func (hb *HeartbeatMonitor) Type(ctx context.Context, raw *heartbeatMonitor.monitor) (graphql2.HeartbeatType, error) {
	return graphql2.HeartbeatType(raw.Type), nil
}
