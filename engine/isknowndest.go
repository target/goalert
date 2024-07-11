package engine

import (
	"context"

	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification"
)

func (n *Engine) IsKnownDest(ctx context.Context, destType notification.DestType, destValue string) (bool, error) {
	d := notification.Dest{Type: destType, Value: destValue}

	b, err := gadb.New(n.b.db).EngineIsKnownDest(ctx, gadb.NullDestV1{Valid: true, DestV1: d.ToDestV1()})
	if err != nil {
		return false, err
	}

	return b.Bool, nil
}
