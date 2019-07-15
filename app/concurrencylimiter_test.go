package app

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPendingList(t *testing.T) {
	list := &pendingList{}

	r1 := list.newReq()
	assert.Equal(t, r1, list.pop())

	r2 := list.newReq()
	assert.Equal(t, r2, list.pop())
	assert.Nil(t, list.head)
	assert.Nil(t, list.tail)

	assert.Nil(t, list.pop())

	r3 := list.newReq()
	r4 := list.newReq()
	r5 := list.newReq()

	assert.True(t, list.remove(r4))
	assert.Nil(t, r4.prev)
	assert.Nil(t, r4.next)
	assert.False(t, list.remove(r4))

	assert.Equal(t, r3, list.pop())
	assert.Equal(t, r5, list.pop())
	assert.Nil(t, list.head)
	assert.Nil(t, list.tail)
}

func TestConcurrencyLimiter(t *testing.T) {
	t.Run("lock count", func(t *testing.T) {
		lim := newConcurrencyLimiter(2, 10)

		ctx, cancel := context.WithCancel(context.Background())
		err := lim.Lock(ctx, "foo")
		assert.Nil(t, err)

		err = lim.Lock(ctx, "foo")
		assert.Nil(t, err)

		err = lim.Lock(ctx, "bar")
		assert.Nil(t, err)

		cancel()
		err = lim.Lock(ctx, "foo")
		assert.Error(t, err, "context canceled")

		lim.Unlock("bar")
		err = lim.Lock(ctx, "foo")
		assert.Error(t, err, "context canceled")

		assert.Panics(t, func() {
			// unlock twice (only locked once)
			lim.Unlock("bar")
		})

		lim.Unlock("foo")
		ctx, cancel = context.WithCancel(context.Background())
		err = lim.Lock(ctx, "foo")
		assert.Nil(t, err)

		cancel()
		assert.Equal(t, 2, lim.lockCount["foo"])
		assert.Equal(t, 0, lim.lockCount["bar"])
	})

	t.Run("queue", func(t *testing.T) {
		lim := newConcurrencyLimiter(1, 3)
		ctx := context.Background()

		gotLockCh := make(chan struct{})
		lock := func() { lim.Lock(ctx, "foo"); gotLockCh <- struct{}{} }
		for i := 0; i < 4; i++ {
			go lock()
		}
		<-gotLockCh
		assert.Equal(t, 1, lim.lockCount["foo"])
		time.Sleep(time.Millisecond) // ensure other goroutines have a chance to fill the queue
		assert.Error(t, lim.Lock(ctx, "foo"))
		assert.Equal(t, 3, lim.queueCount["foo"])

		lim.Unlock("foo")
		go lock()

		<-gotLockCh
		lim.Unlock("foo")

		<-gotLockCh
		lim.Unlock("foo")

		<-gotLockCh
		lim.Unlock("foo")

		<-gotLockCh
		lim.Unlock("foo")

		assert.Equal(t, 0, lim.lockCount["foo"])
		assert.Equal(t, 0, lim.queueCount["foo"])

		assert.Nil(t, lim.Lock(ctx, "foo"))
	})
}
