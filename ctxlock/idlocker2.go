package ctxlock

// import (
// 	"context"
// 	"sync"
// )

// type Queue struct {
// 	n int

// 	_id  int64
// 	wait []chan struct{}
// 	c    sync.Cond
// }

// type IDLocker2[K comparable] struct {
// 	cfg Config
// 	m   sync.Map
// 	p   sync.Pool
// }

// func NewIDLocker2[K comparable](cfg Config) *IDLocker2[K] {
// 	return &IDLocker2[K]{
// 		cfg: cfg,
// 		p: sync.Pool{
// 			New: func() any { return make() },
// 		},
// 	}
// }

// func (l2 *IDLocker2[K]) Lock(ctx context.Context, id K) error {
// 	val, load := l2.m.LoadOrStore(id, 1)

// 	return nil
// }

// func (l2 *IDLocker2[K]) Unlock(id K) {
// 	val, ok := l2.m.Load(id)
// 	if !ok {
// 		panic("unlock of unheld lock")
// 	}
// }
