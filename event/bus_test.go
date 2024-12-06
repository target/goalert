package event

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBus(t *testing.T) {
	b := NewBus(slog.Default())

	type testTx struct {
		ID int
	}

	var batch, batchTx []int
	var sentTx testTx
	OnEachBatch(b, func(ctx context.Context, data []int) error {
		batch = data
		return nil
	})
	OnEachBatchTx(b, func(ctx context.Context, tx testTx, data []int) error {
		batchTx = data
		sentTx = tx
		return nil
	})

	Send(context.Background(), b, 3)
	require.Equal(t, []int{3}, batch)
	SendMany(context.Background(), b, []int{1, 2, 3})
	require.Equal(t, []int{1, 2, 3}, batch)

	SendTx(context.Background(), b, testTx{ID: 1}, 3)
	require.Equal(t, []int{3}, batchTx)
	require.Equal(t, testTx{ID: 1}, sentTx)
	SendManyTx(context.Background(), b, testTx{ID: 2}, []int{1, 2, 3})
	require.Equal(t, []int{1, 2, 3}, batchTx)
	require.Equal(t, testTx{ID: 2}, sentTx)
}
