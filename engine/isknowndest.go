package engine

import (
	"context"
	"database/sql"
	"errors"

	"github.com/target/goalert/notification"
)

func (n *Engine) IsKnownDest(ctx context.Context, destType notification.DestType, destValue string) (bool, error) {
	var isKnown bool
	var err error
	if destType.IsUserCM() {
		err = n.b.validCM.QueryRowContext(ctx, destType.CMType(), destValue).Scan(&isKnown)
	} else {
		err = n.b.validNC.QueryRowContext(ctx, destType.NCType(), destValue).Scan(&isKnown)
	}
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}

	return isKnown, err
}
