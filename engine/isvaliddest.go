package engine

import (
	"context"

	"github.com/target/goalert/notification"
)

func (n *Engine) IsValidDest(ctx context.Context, destType notification.DestType, destValue string) (bool, error) {
	return true, nil
}
