package graphqlapp

import (
	context "context"
	"database/sql"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/heartbeat"
)

type Heartbeat App

func (a *App) Heartbeat() graphql2.HeartbeatResolver { return (*Heartbeat)(a) }

func (q *Query) Heartbeat(ctx context.Context, id string) (*heartbeat.Heartbeat, error) {
	return q.FindAllByService(ctx, id)
}

func (m *Mutation) CreateHeartbeat(ctx context.Context, input graphql2.CreateHeartbeatInput) (beat *heartbeat.Heartbeat, err error) {
	var serviceID string
	if input.ServiceID != nil {
		serviceID = *input.ServiceID
	}
	err = withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		beat = &heartbeat.Heartbeat{
			ServiceID:         serviceID,
			Name:              input.Name,
			HeartbeatInterval: input.HeartbeatInterval,
			LastState:         input.LastState,
			// lastHeartbeat not required - include here?
		}
		beat, err = m.DB.hbStore.CreateTx(ctx, tx, beat)
		return err
	})
	return beat, err
}

func (beat *Heartbeat) Type(ctx context.Context, raw *heartbeat.Heartbeat) (graphql2.HeartbeatType, error) {
	return graphql2.HeartbeatType(raw.Type), nil
}
