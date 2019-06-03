package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrencyLimiter(t *testing.T) {
	lim := newConcurrencyLimiter(2)

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
}
