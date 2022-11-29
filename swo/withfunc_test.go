package swo

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test 1 immediately return error
// Test 2 call call-back fn, should block until cancel()
func TestBegin(t *testing.T) {
	wf1 := NewWithFunc(func(ctx context.Context, fn func(struct{})) error {
		return errors.New("expected error")
	})

	_, err := wf1.Begin(context.Background())
	assert.Error(t, err)

}
