package graphqlapp

import (
	context "context"
	"database/sql"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/heartbeat"
)

type HeartbeatMonitor App

func (a *App) HeartbeatMonitor() graphql2.HeartbeatResolver { return (*HeartbeatMonitor)(a) }

func (q *Query) HeartbeatMonitor(ctx context.Context, id string) (*heartbeat.Monitor, error) {
	return (*App)(q).FindOneHeartbeatMonitor(ctx, id)
}

func (m *Mutation) CreateHeartbeatMonitor(ctx context.Context, input graphql2.CreateHeartbeatMonitorInput) (*heartbeat.Monitor, error) {
	hb := &heartbeat.Monitor{
		ServiceID:      input.ServiceID,
		Name:           input.Name,
		TimeoutMinutes: input.TimeoutMinutes,
	}
	var heartbeatMonitor *heartbeat.Monitor
	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		heartbeatMonitor, err := m.HeartbeatStore.CreateTx(ctx, tx, hb)
		return err
	})
	return heartbeatMonitor, err
}

func (m *Mutation) UpdateHeartbeatMonitor(ctx context.Context, input graphql2.UpdateHeartbeatMonitorInput) (bool, error) {
	hb := &heartbeat.Monitor{
		ID: input.ID,
	}
	if input.Name != nil {
		hb.Name = *input.Name
	}
	if input.TimeoutMinutes != nil {
		hb.TimeoutMinutes = *input.TimeoutMinutes
	}
	var result bool
	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		result := m.HeartbeatStore.UpdateTx(ctx, tx, hb)
		return result
	})
	return result, err
}
