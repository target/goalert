package swo

import "context"

type Syncer interface {
	Reset(context.Context) error
	Setup(context.Context) error
	Sync(ctx context.Context, progress func(float64)) error
}
