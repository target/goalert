package app

type concurrencyLimiter struct {
	max     int
	count   map[string]int
	pending map[string]chan struct{}

	mx sync.Mutex
}

func newConcurrencyLimiter(max int) *concurrencyLimiter {
	return &concurrencyLimiter{
		max:     max,
		count:   make(map[string]int, 100),
		pending: make(map[string]chan struct{}),
	}
}

func (l *concurrencyLimiter) Lock(ctx context.Context, id string) error {
	for {
		l.mx.Lock()
		n := l.count[id]
		if n < l.max {
			l.count[id] = n + 1
			l.mx.Unlock()
			return nil
		}

		ch := l.pending[id]
		if ch == nil {
			ch = make(chan struct{})
			l.pending[id] = ch
		}
		l.mx.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ch:
		}
	}
}

func (l *concurrencyLimiter) Unlock(id string) {
	l.mx.Lock()
	n := l.count[id]
	n--
	if n == 0 {
		delete(l.count, id)
	} else {
		l.count[id] = n
	}
	ch := l.pending[id]
	if ch != nil {
		delete(l.pending, id)
		close(ch)
	}
	l.mx.Unlock()
}
