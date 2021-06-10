package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiLock_TryLock(t *testing.T) {
	l := newMultiLock(1, 1)
	ctx := context.Background()
	assert.NoError(t, l.Lock(ctx))

	wait, err := l.TryLock()
	assert.NoError(t, err)

	go wait(ctx)
	assert.False(t, l.Unlock())
	assert.True(t, l.Unlock())
}

func TestMultiLock(t *testing.T) {
	l := newMultiLock(10, 10)

	err := l.Lock(context.Background())
	require.NoError(t, err)

	for i := 0; i < 19; i++ {
		wait, err := l.TryLock()
		require.NoError(t, err)
		if wait != nil {
			go wait(context.Background())
		}
	}

	for i := 0; i < 20; i++ {
		assert.False(t, l.Unlock())
		wait, err := l.TryLock()
		require.NoError(t, err)
		go wait(context.Background())
	}

	for i := 0; i < 19; i++ {
		assert.False(t, l.Unlock())
	}

	assert.True(t, l.Unlock())
}

func TestConcurrencyLimiter(t *testing.T) {
	t.Run("lock count", func(t *testing.T) {
		lim := newConcurrencyLimiter(2, 0)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		assert.NoError(t, lim.Lock(ctx, "foo"))
		assert.NoError(t, lim.Lock(ctx, "foo"))
		assert.Error(t, lim.Lock(ctx, "foo"))
		assert.NoError(t, lim.Lock(ctx, "bar"))
		lim.Unlock("foo")
		assert.NoError(t, lim.Lock(ctx, "foo"))
	})

	t.Run("cancelation", func(t *testing.T) {
		lim := newConcurrencyLimiter(1, 1)
		ctx, cancel := context.WithCancel(context.Background())
		assert.NoError(t, lim.Lock(ctx, "foo"))
		cancel()
		assert.Error(t, lim.Lock(ctx, "foo"))

		lim.Unlock("foo")

		// empty limiter should return lock instantly, even with canceled context (required behavior for empty/LRU cleanup logic)
		//
		// context only comes into play if in queue/waiting.
		assert.NoError(t, lim.Lock(ctx, "foo"))
	})

	t.Run("queue", func(t *testing.T) {
		lim := newConcurrencyLimiter(1, 2)

		ctx, cancel := context.WithCancel(context.Background())
		assert.NoError(t, lim.Lock(ctx, "foo"))

		errCh := make(chan error, 3)
		go func() { errCh <- lim.Lock(ctx, "foo") }()
		go func() { errCh <- lim.Lock(ctx, "foo") }()
		go func() { errCh <- lim.Lock(ctx, "foo") }()

		// one too many for the config
		assert.EqualError(t, <-errCh, errQueueFull.Error())

		// next up in queue should work
		lim.Unlock("foo")
		assert.NoError(t, <-errCh)

		// cancel while last one is in queue
		cancel()
		assert.Error(t, <-errCh)
	})
}
