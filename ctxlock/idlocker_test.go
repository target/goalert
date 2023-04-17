package ctxlock_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/ctxlock"
)

func TestIDLocker_NoQueue(t *testing.T) {
	assert.Panics(t, func() {
		ctxlock.NewIDLocker[string](ctxlock.Config{MaxHeld: -1})
	}, "MaxHeld must be >= 0")

	l := ctxlock.NewIDLocker[string](ctxlock.Config{})

	ctx := context.Background()

	err := l.Lock(ctx, "foo")
	require.NoError(t, err, "first lock should work")

	err = l.Lock(ctx, "foo")
	require.ErrorIs(t, err, ctxlock.ErrQueueFull, "second lock should fail with ErrQueueFull")

	err = l.Lock(ctx, "bar")
	require.NoError(t, err, "lock for different id should work")

	l.Unlock("foo")
	err = l.Lock(ctx, "foo")
	require.NoError(t, err, "third lock should work")

	l.Unlock("foo")
	require.Panics(t, func() {
		l.Unlock("foo") // too many unlocks
	})
}

func TestIDLocker_Context(t *testing.T) {
	l := ctxlock.NewIDLocker[string](ctxlock.Config{MaxWait: 1})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := l.Lock(ctx, "foo")
	require.NoError(t, err, "first lock should work")

	ch := make(chan error, 2)
	go func() { ch <- l.Lock(ctx, "foo") }()
	go func() { ch <- l.Lock(ctx, "foo") }()

	err = <-ch
	require.ErrorIs(t, err, ctxlock.ErrQueueFull, "second lock should fail with ErrQueueFull")
	l.Unlock("foo")

	err = <-ch
	require.NoError(t, err, "third lock should work")

	cancel()
	err = l.Lock(ctx, "foo")
	require.ErrorIs(t, err, context.Canceled, "lock should fail with context canceled")
}

func TestIDLocker_Timeout(t *testing.T) {
	l := ctxlock.NewIDLocker[string](ctxlock.Config{MaxWait: 1, Timeout: time.Millisecond})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := l.Lock(ctx, "foo")
	require.NoError(t, err, "first lock should work")

	ch := make(chan error, 1)
	go func() { ch <- l.Lock(ctx, "foo") }()

	err = <-ch
	require.ErrorIs(t, err, ctxlock.ErrTimeout, "third lock should fail with ErrTimeout")
}

func TestIDLocker_CancelQueue(t *testing.T) {
	l := ctxlock.NewIDLocker[string](ctxlock.Config{MaxWait: 10, Timeout: time.Millisecond})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := l.Lock(ctx, "foo")
	require.NoError(t, err, "first lock should work")

	ch := make(chan error)
	nnn, nCancel := context.WithCancel(ctx)
	// We need to test that canceling items in the queue works.
	// Our nnn will be the one we cancel. We're going to fill the queue
	// with these, then Unlock a few times to make new space, fill it with
	// ones waiting on ctx, then do again, then cancel nnn.
	//
	// Our queue will look like this:
	// [nnn, nnn, nnn, ctx, ctx, nnn, ctx, ctx, ctx, ctx]
	//
	// The only way we can be sure the queue is full, is to push items until
	// we get ErrQueueFull. Doing so will allow us to test the logic of the
	// concurrent queue, without relying on timing.
	//
	// Note: all the `go func() { ... }` calls can run in any order, so we
	// use the ErrQueueFull to guarantee that we've filled/processed all the
	// items we want to test before pushing more.
	for i := 0; i < 11; i++ {
		go func() { ch <- l.Lock(nnn, "foo") }()
	}
	require.ErrorIs(t, <-ch, ctxlock.ErrQueueFull) // queue overflow
	// Queue now looks like this (all nnn):
	// [nnn, nnn, nnn, nnn, nnn, nnn, nnn, nnn, nnn, nnn, nnn]

	// We need to make space for the two ctx items (position 3 and 4).
	l.Unlock("foo")
	l.Unlock("foo")
	require.NoError(t, <-ch)
	require.NoError(t, <-ch)

	// Attempt to push 3 ctx items.
	go func() { ch <- l.Lock(ctx, "foo") }()
	go func() { ch <- l.Lock(ctx, "foo") }()
	go func() { ch <- l.Lock(ctx, "foo") }()
	require.ErrorIs(t, <-ch, ctxlock.ErrQueueFull) // queue overflow on the last
	// Queue now should look like this:
	// [nnn, nnn, nnn, nnn, nnn, nnn, nnn, nnn, ctx, ctx]

	// Make one space, push two nnn items.
	// One of them will randomly end up in the last spot, and the other
	// will be rejected with ErrQueueFull.
	l.Unlock("foo")
	require.NoError(t, <-ch)
	go func() { ch <- l.Lock(nnn, "foo") }()
	go func() { ch <- l.Lock(nnn, "foo") }()
	require.ErrorIs(t, <-ch, ctxlock.ErrQueueFull) // queue overflow
	// Queue now looks like this:
	// [nnn, nnn, nnn, nnn, nnn, nnn, nnn, ctx, ctx, nnn]

	// We want that last nnn to move from position 9 to position 5.

	l.Unlock("foo")
	l.Unlock("foo")
	l.Unlock("foo")
	l.Unlock("foo")
	require.NoError(t, <-ch)
	require.NoError(t, <-ch)
	require.NoError(t, <-ch)
	require.NoError(t, <-ch)

	// Queue now looks like this:
	// [nnn, nnn, nnn, ctx, ctx, nnn]
	for i := 0; i < 5; i++ {
		go func() { ch <- l.Lock(ctx, "foo") }()
	}
	require.ErrorIs(t, <-ch, ctxlock.ErrQueueFull) // queue overflow
	// Queue now looks like this:
	// [nnn, nnn, nnn, ctx, ctx, nnn, ctx, ctx, ctx, ctx]

	// Now we can cancel nnn, and ensure we get our 4 canceled items.
	nCancel()
	for i := 0; i < 4; i++ {
		require.ErrorIs(t, <-ch, context.Canceled)
	}

	// At this point, we should have 6 items in the queue.
	// We can now unlock 6 times, and ensure we get 6 items.
	for i := 0; i < 6; i++ {
		l.Unlock("foo")
		require.NoError(t, <-ch)
	}

	l.Unlock("foo") // original lock

	// We should now have an empty queue, and no held lock.
	require.Panics(t, func() { l.Unlock("foo") }, "unlocking an empty queue should panic")
}

func TestIDLocker_Unlock_Abandoned(t *testing.T) {
	l := ctxlock.NewIDLocker[string](ctxlock.Config{
		MaxWait: 1,
		Timeout: time.Second, // in case of bug, we don't want to wait forever
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := l.Lock(ctx, "foo")
	require.NoError(t, err, "first lock should work")

	test := func() {
		nnn, cancel := context.WithCancel(ctx)
		ch := make(chan error, 1)
		go func() { ch <- l.Lock(nnn, "foo") }()
		go func() { ch <- l.Lock(nnn, "foo") }()
		require.ErrorIs(t, <-ch, ctxlock.ErrQueueFull, "queue should be full")

		cancel()
		l.Unlock("foo")

		if <-ch != nil {
			// lost the race, re-lock foo
			err := l.Lock(ctx, "foo")
			require.NoError(t, err, "re-lock should work")
		}
	}

	for i := 0; i < 1000; i++ {
		test()
	}

	l.Unlock("foo") // original lock
	assert.Panics(t, func() { l.Unlock("foo") }, "unlocking an empty queue should panic")
}

func BenchmarkIDLocker_Sequential(b *testing.B) {
	l := ctxlock.NewIDLocker[struct{}](ctxlock.Config{})
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		err := l.Lock(ctx, struct{}{})
		if err != nil {
			b.Fatal(err)
		}
		l.Unlock(struct{}{})
	}
}

func BenchmarkIDLocker_Sequential_Cardinality(b *testing.B) {
	l := ctxlock.NewIDLocker[int64](ctxlock.Config{})
	ctx := context.Background()
	var n int64
	for i := 0; i < b.N; i++ {
		err := l.Lock(ctx, n)
		if err != nil {
			b.Fatal(err)
		}
		n++
		if n > 100 {
			l.Unlock(n - 100)
		}
	}
}

func BenchmarkIDLocker_Concurrent(b *testing.B) {
	l := ctxlock.NewIDLocker[struct{}](ctxlock.Config{MaxWait: -1})
	ctx := context.Background()

	b.SetParallelism(1000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := l.Lock(ctx, struct{}{})
			require.NoError(b, err)
			l.Unlock(struct{}{})
		}
	})
}

func BenchmarkIDLocker_Concurrent_Cardinality(b *testing.B) {
	l := ctxlock.NewIDLocker[int64](ctxlock.Config{MaxWait: 1})
	ctx := context.Background()

	b.SetParallelism(1000)
	var n int64
	b.RunParallel(func(pb *testing.PB) {
		id := atomic.AddInt64(&n, 1)
		ch := make(chan error, 1)
		for pb.Next() {
			err := l.Lock(ctx, id)
			require.NoError(b, err)
			go func() { ch <- l.Lock(ctx, id) }()
			l.Unlock(id)
			require.NoError(b, <-ch)
			l.Unlock(id)
		}
	})
}
