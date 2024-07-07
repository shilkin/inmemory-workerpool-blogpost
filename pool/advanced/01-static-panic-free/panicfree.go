package _1_static_panic_free

import (
	"context"
	"errors"
	"sync/atomic"
)

// Task is public Task interface
// that receives task-scope context.
type Task func(taskCtx context.Context)

type Pool struct {
	// taskC is a task queue
	taskC chan func()

	// isStopping shows that the pool is stopping
	isStopping atomic.Bool

	// ctx is a pool context that signals that the pool is stopping
	ctx context.Context

	// cancel cancels pool context
	cancel func()
}

// NewPool creates worker pool of
// size goroutines running and ready
// to execute tasks.
func NewPool(size int) *Pool {
	// Buffered channel that works as a task queue
	taskC := make(chan func(), size)

	poolCtx, cancel := context.WithCancel(context.Background())

	// Run limited number of workers
	for i := 0; i < size; i++ {
		go func() {
			for {
				select {
				case task, ok := <-taskC:
					// safety check
					if !ok {
						return
					}

					task()
				case <-poolCtx.Done():
					// pool is stopping
					return
				}
			}
		}()
	}

	return &Pool{
		taskC:  taskC,
		ctx:    poolCtx,
		cancel: cancel,
	}
}

func (p *Pool) Enqueue(ctx context.Context, task Task) error {
	// do not allow enqueueing when pool is stopping
	if p.isStopping.Load() {
		return errors.New("pool is stopping")
	}

	// note that we make the task context non-cancellable
	// to avoid cascading parent context cancellation
	taskCtx := context.WithoutCancel(ctx)

	// note that all in-flight Enqueue at this point may
	// still send tasks to taskC because select operator
	// does not determine the order of cases;
	//
	// we can think of the second select but this complicates
	// the code and doesn't address the issue in full so let's
	// treat it as an edge case
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.ctx.Done():
		return errors.New("pool is stopping")
	case p.taskC <- func() { task(taskCtx) }:
		return nil
	}
}

// Stop stops all the workers.
func (p *Pool) Stop() {
	// set the stopping flag
	p.isStopping.Store(true)

	// send a signal to those who waits
	p.cancel()
}
