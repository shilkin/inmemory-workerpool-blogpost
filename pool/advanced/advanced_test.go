package advanced_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	static "github.com/shilkin/inmemory-workerpool-blogpost/pool/advanced/00-static"
	panicfree "github.com/shilkin/inmemory-workerpool-blogpost/pool/advanced/01-static-panic-free"
	waitable "github.com/shilkin/inmemory-workerpool-blogpost/pool/advanced/02-static-waitable"
	timeoutwaitable "github.com/shilkin/inmemory-workerpool-blogpost/pool/advanced/03-static-timeout-waitable"
	taskstopable "github.com/shilkin/inmemory-workerpool-blogpost/pool/advanced/04-static-task-stopable"
	dynamic "github.com/shilkin/inmemory-workerpool-blogpost/pool/advanced/06-dynamic"
)

func TestStaticPool(t *testing.T) {
	t.Run("enqueue normally", func(t *testing.T) {
		size := 10
		pool := static.NewPool(size)

		// give it a time to start before stopping
		time.Sleep(5 * time.Millisecond)

		var taskExecuted atomic.Int32

		for i := 0; i < size; i++ {
			err := pool.Enqueue(context.Background(), func(_ context.Context) {
				taskExecuted.Add(1)
			})
			require.NoError(t, err)
		}

		pool.Stop()

		// we are not waiting for each task to be done
		// so the actual number of executed tasks may be
		// less then it was scheduled
		require.LessOrEqual(t, int(taskExecuted.Load()), size)
	})

	t.Run("enqueue panics after stop", func(t *testing.T) {
		pool := static.NewPool(1)
		pool.Stop()

		require.Panics(t, func() {
			_ = pool.Enqueue(context.Background(), func(_ context.Context) {})
		})
	})
}

func TestPanicFree(t *testing.T) {
	pool := panicfree.NewPool(1)
	pool.Stop()

	require.NotPanics(t, func() {
		_ = pool.Enqueue(context.Background(), func(_ context.Context) {})
	})
}

func TestWaitablePool(t *testing.T) {
	size := 10
	pool := waitable.NewPool(size)

	// give it a time to start before stopping
	time.Sleep(5 * time.Millisecond)

	var taskExecuted atomic.Int32

	for i := 0; i < size; i++ {
		err := pool.Enqueue(context.Background(), func(_ context.Context) {
			taskExecuted.Add(1)
		})
		require.NoError(t, err)
	}

	pool.Stop()

	// we are waiting for each task to be done
	// so the actual number of executed tasks
	// must be equal to the number of scheduled tasks
	require.Equal(t, size, int(taskExecuted.Load()))
}

func TestTimeoutPool(t *testing.T) {
	t.Run("stop normally", func(t *testing.T) {
		pool := timeoutwaitable.NewPool(1)

		err := pool.Enqueue(context.Background(), func(_ context.Context) {
			time.Sleep(1 * time.Millisecond)
		})
		require.NoError(t, err)

		err = pool.Stop(10 * time.Millisecond)
		require.NoError(t, err)
	})

	t.Run("stop timeout", func(t *testing.T) {
		pool := timeoutwaitable.NewPool(1)

		err := pool.Enqueue(context.Background(), func(_ context.Context) {
			time.Sleep(1 * time.Minute)
		})
		require.NoError(t, err)

		// give the single static worker
		// to take the task from the queue
		time.Sleep(10 * time.Millisecond)

		err = pool.Stop(2 * time.Millisecond)
		require.Error(t, err)
	})
}

func TestTaskStopable(t *testing.T) {
	pool := taskstopable.NewPool(1)

	// give it a time to start before stopping
	time.Sleep(5 * time.Millisecond)

	var taskStopped atomic.Bool
	var taskExecuted atomic.Bool

	err := pool.Enqueue(context.Background(), func(poolCtx, taskCtx context.Context) {
		// wait for task being stopped or executed
		select {
		// task timeout
		case <-taskCtx.Done(): // noop

		// pool is stopping
		case <-poolCtx.Done():
			taskStopped.Store(true)

		// execute task
		case <-time.After(1 * time.Second):
			taskExecuted.Store(true)
		}
	})
	require.NoError(t, err)

	err = pool.Stop(10 * time.Millisecond)
	require.NoError(t, err)
	require.True(t, taskStopped.Load())
	require.False(t, taskExecuted.Load())
}

func TestDynamicPool(t *testing.T) {
	t.Run("wrong pool size", func(t *testing.T) {
		minSize, maxSize := 10, 5

		_, err := dynamic.NewPool(minSize, maxSize)
		require.Error(t, err)
	})

	t.Run("dynamic workers", func(t *testing.T) {
		minSize, maxSize := 1, 3

		pool, err := dynamic.NewPool(minSize, maxSize)
		require.NoError(t, err)
		require.NotNil(t, pool)

		// schedule minSize+1 slow tasks
		// to make one handling and one
		// queued task in the channel
		for i := 0; i < minSize+1; i++ {
			err = pool.Enqueue(context.Background(), func(poolCtx, taskCtx context.Context) {
				select {
				case <-taskCtx.Done():
					return
				case <-poolCtx.Done():
					return
				case <-time.After(time.Minute):
					return
				}
			})
			require.NoError(t, err)

			// give the single static worker
			// to take the task from the queue
			time.Sleep(10 * time.Millisecond)
		}

		var taskExecuted atomic.Int32

		// schedule extra maxSize-minSize fast tasks dynamically
		for i := 0; i < maxSize-minSize; i++ {
			// set enqueueing timeout to check that
			// we don't wait for scheduling because
			// of all static workers are busy
			timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
			defer cancel()

			err = pool.Enqueue(timeoutCtx, func(_, _ context.Context) {
				taskExecuted.Add(1)
			})

			require.NoError(t, err)
		}

		err = pool.Stop(10 * time.Millisecond)
		require.NoError(t, err)

		require.Equal(t, maxSize-minSize, int(taskExecuted.Load()))
	})
}
