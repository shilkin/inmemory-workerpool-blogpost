package service

import (
	"context"
	"github.com/stretchr/testify/require"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolServiceEnqueue(t *testing.T) {
	ps := NewPoolService(3)

	var counter int32

	task := func(ctx context.Context) {
		atomic.AddInt32(&counter, 1)
	}

	ctx := context.Background()

	for i := 0; i < 10; i++ {
		err := ps.Enqueue(ctx, task)
		require.NoError(t, err, "failed to enqueue task")
	}

	ps.Stop()

	require.Equal(t, int32(10), counter, "expected counter to be 10")
}

func TestPoolServiceStop(t *testing.T) {
	ps := NewPoolService(3)
	ps.Stop()

	ctx := context.Background()
	err := ps.Enqueue(ctx, func(ctx context.Context) {})
	require.Error(t, err, "expected error when enqueueing task to stopped pool")
}

func TestPoolServiceGracefulShutdown(t *testing.T) {
	ps := NewPoolService(3)

	var counter int32

	task := func(ctx context.Context) {
		time.Sleep(100 * time.Millisecond)
		atomic.AddInt32(&counter, 1)
	}

	ctx := context.Background()

	for i := 0; i < 30; i++ {
		err := ps.Enqueue(ctx, task)
		require.NoError(t, err, "failed to enqueue task")
	}

	time.Sleep(50 * time.Millisecond)
	ps.Stop()

	time.Sleep(3100 * time.Millisecond)

	require.Equal(t, int32(30), counter, "expected counter to be 30")
}
