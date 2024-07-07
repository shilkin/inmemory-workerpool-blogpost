package _0_static

import (
	"context"
)

// Task is public Task interface
// that receives task-scope context.
type Task func(taskCtx context.Context)

type Pool struct {
	// taskC is a task queue
	taskC chan func()
}

// NewPool creates worker pool of
// size goroutines running and ready
// to execute tasks.
func NewPool(size int) *Pool {
	// Buffered channel that works as a task queue
	taskC := make(chan func(), size)

	// Run limited number of workers
	for i := 0; i < size; i++ {
		go func() {
			// Read from the task queue
			for task := range taskC {
				// Execute the task
				task()
			}
		}()
	}

	return &Pool{taskC: taskC}
}

func (p *Pool) Enqueue(ctx context.Context, task Task) error {
	// note that we make the task context non-cancellable
	// to avoid cascading parent context cancellation
	taskCtx := context.WithoutCancel(ctx)

	select {
	case <-ctx.Done():
		return ctx.Err()

	// this panics when writing to closed channel
	case p.taskC <- func() { task(taskCtx) }:
		return nil
	}
}

// Stop stops all the workers.
func (p *Pool) Stop() {
	close(p.taskC)
}
