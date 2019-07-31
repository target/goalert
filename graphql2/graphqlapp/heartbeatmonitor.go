package graphqlapp

import (
	context "context"
	"database/sql"
	"net/url"
	"time"

	"github.com/target/goalert/config"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/heartbeat"
)

type HeartbeatMonitor App

func (a *App) HeartbeatMonitor() graphql2.HeartbeatMonitorResolver { return (*HeartbeatMonitor)(a) }

func (a *HeartbeatMonitor) TimeoutMinutes(ctx context.Context, hb *heartbeat.Monitor) (int, error) {
	return int(hb.Timeout / time.Minute), nil
}
func (a *HeartbeatMonitor) Href(ctx context.Context, hb *heartbeat.Monitor) (string, error) {
	cfg := config.FromContext(ctx)
	return cfg.CallbackURL("/api/v2/heartbeat/" + url.PathEscape(hb.ID)), nil
}

func (q *Query) HeartbeatMonitor(ctx context.Context, id string) (*heartbeat.Monitor, error) {
	return (*App)(q).FindOneHeartbeatMonitor(ctx, id)
}

func (m *Mutation) CreateHeartbeatMonitor(ctx context.Context, input graphql2.CreateHeartbeatMonitorInput) (*heartbeat.Monitor, error) {
	hb := &heartbeat.Monitor{
		ServiceID: input.ServiceID,
		Name:      input.Name,
		Timeout:   time.Duration(input.TimeoutMinutes) * time.Minute,
	}
	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		hb, err = m.HeartbeatStore.CreateTx(ctx, tx, hb)
		return err
	})
	return hb, err
}

func (m *Mutation) UpdateHeartbeatMonitor(ctx context.Context, input graphql2.UpdateHeartbeatMonitorInput) (bool, error) {
	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		hb, err := m.HeartbeatStore.FindOneTx(ctx, tx, input.ID)
		if err != nil {
			return err
		}
		if input.Name != nil {
			hb.Name = *input.Name
		}
		if input.TimeoutMinutes != nil {
			hb.Timeout = time.Duration(*input.TimeoutMinutes) * time.Minute
		}

		return m.HeartbeatStore.UpdateTx(ctx, tx, hb)
	})
	return err == nil, err
}
