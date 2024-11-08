package signalmgr

import (
	"context"

	"github.com/target/goalert/engine/processinglock"
)

var _ processinglock.Setupable = &DB{}

func (db *DB) Setup(ctx context.Context, args processinglock.SetupArgs) error {
	return nil
}
