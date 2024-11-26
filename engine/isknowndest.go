package engine

import (
	"context"

	"github.com/target/goalert/gadb"
)

func (n *Engine) IsKnownDest(ctx context.Context, dest gadb.DestV1) (bool, error) {
	b, err := gadb.NewCompat(n.b.db).EngineIsKnownDest(ctx, gadb.NullDestV1{Valid: true, DestV1: dest})
	if err != nil {
		return false, err
	}

	return b.Bool, nil
}
