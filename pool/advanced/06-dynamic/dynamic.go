package _6_dynamic

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Task is public Task interface
// that receives task-scope context.
type Task func(poolCtx, taskCtx context.Context)

type Pool struct {
	// taskC is a task queue
	taskC chan func()

	// wg is a WaitGroup for waiting all workers to finish
	wg *sync.WaitGroup

	// ctx is a pool context that signals that the pool is stopping
	ctx context.Context

	// cancel cancels pool context
	cancel func()

	// isStopping shows that the pool is stopping
	isStopping atomic.Bool

	// dynamicPoolC is a semaphore for running dynamic workers
	dynamicPoolC chan struct{}

	// stopMutex protects wait group when stopping the pool
	stopMutex sync.Mutex
}

// NewPool creates worker pool of
// minSize goroutines running and ready
// to execute tasks.
//
// If no static workers available it is
// able to extend the pool up to maxSize workers
// which quit after task execution.
func NewPool(minSize, maxSize int) (*Pool, error) {
	if minSize >= maxSize {
		return nil, fmt.Errorf("wrong pool size: minSize %d >= maxSize %d", minSize, maxSize)
	}

	taskC := make(chan func(), minSize)
	wg := sync.WaitGroup{}
	poolCtx, cancel := context.WithCancel(context.Background())

	// Run limited number of workers
	for i := 0; i < minSize; i++ {
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
		taskC:        taskC,
		ctx:          poolCtx,
		cancel:       cancel,
		wg:           &wg,
		dynamicPoolC: make(chan struct{}, maxSize-minSize),
	}, nil
}

func (p *Pool) Enqueue(ctx context.Context, task Task) error {
	if p.isStopping.Load() {
		return errors.New("pool is stopping")
	}

	taskCtx := context.WithoutCancel(ctx)
	poolCtx := p.ctx

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.ctx.Done():
		return errors.New("pool is stopping")
	case p.taskC <- func() { task(poolCtx, taskCtx) }:
		return nil
	default:
		// if no static workers available then
		// try to schedule dynamic worker
		return p.enqueueDynamic(taskCtx, task)
	}
}

func (p *Pool) enqueueDynamic(taskCtx context.Context, task Task) error {
	select {
	case <-taskCtx.Done():
		return taskCtx.Err()
	case <-p.ctx.Done():
		return errors.New("pool is stopping")
	// acquire worker
	case p.dynamicPoolC <- struct{}{}:
		// note that at this point Stop can be called,
		// acquire mutex to protect wait group counter
		// to be incremented after Wait being called in
		// Stop function
		p.stopMutex.Lock() // <-

		if p.isStopping.Load() {
			p.stopMutex.Unlock() // <-

			return errors.New("pool is stopping")
		}

		// this shouldn't happen after Wait
		p.wg.Add(1)

		p.stopMutex.Unlock() // <-

		// schedule dynamic worker
		go func() {
			// release worker
			defer func() { <-p.dynamicPoolC }()

			defer p.wg.Done()

			task(p.ctx, taskCtx)
		}()
	}

	return nil
}

// Stop stops all the workers and waits them to finish for a certain time.
func (p *Pool) Stop(timeout time.Duration) error {
	// to avoid calling wg.Wait before wg.Add in dynamicEnqueue
	// we have to introduce a critical section
	//
	// note that in this case using of atomic.Bool is unnecessary,
	// but we leave it just for code similarity,
	// it keeps the code safe even if you remove dynamic workers
	p.stopMutex.Lock()
	defer p.stopMutex.Unlock()

	// now we are quitting the function every time
	// it is already being called
	if !p.isStopping.CompareAndSwap(false, true) {
		return nil
	}

	p.cancel()

	doneC := make(chan struct{})

	go func() {
		p.wg.Wait()
		close(doneC)
	}()

	select {
	case <-doneC:
		return nil
	case <-time.After(timeout):
		return errors.New("stop timeout")
	}
}
