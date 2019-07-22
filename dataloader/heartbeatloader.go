package dataloader

import (
	"context"

	"github.com/target/goalert/heartbeat"
)

type HeartbeatLoader struct {
	*loader
	store heartbeat.Store
}

func NewHeartbeatLoader(ctx context.Context, store heartbeat.Store) *HeartbeatLoader {
	p := &HeartbeatLoader{
		store: store,
	}
	p.loader = newLoader(ctx, loaderConfig{})
	return p
}
