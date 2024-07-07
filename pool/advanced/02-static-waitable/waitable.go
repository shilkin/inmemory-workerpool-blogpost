package _2_static_waitable

import (
	"context"
	"errors"
	"sync"
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

	// wg is a WaitGroup for waiting all workers to finish
	wg *sync.WaitGroup
}

// NewPool creates worker pool of
// size goroutines running and ready
// to execute tasks.
func NewPool(size int) *Pool {
	taskC := make(chan func(), size)
	wg := sync.WaitGroup{}
	poolCtx, cancel := context.WithCancel(context.Background())

	// Run limited number of workers
	for i := 0; i < size; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

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
		wg:     &wg,
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
	// let's treat it as an edge case
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
	p.isStopping.Store(true)

	p.cancel()

	p.wg.Wait()
}
