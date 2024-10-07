package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type Task func(poolCtx, taskctx context.Context)

type WorkerPool struct {
	queue chan func()
	count int
	stop  chan struct{}
	wg    sync.WaitGroup

	isStopping atomic.Bool

	poolCtx    context.Context
	cancelTask func()
}

func NewWorkerPool(count int) *WorkerPool {
	poolCtx, cancelTask := context.WithCancel(context.Background())

	pool := &WorkerPool{
		queue:      make(chan func(), count),
		count:      count,
		stop:       make(chan struct{}),
		poolCtx:    poolCtx,
		cancelTask: cancelTask,
	}

	for i := 0; i < count; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}

	return pool
}

func (p *WorkerPool) worker() {
	defer p.wg.Done()
	for {
		select {
		case <-p.stop:
			return
		case task := <-p.queue:
			task()
		}
	}
}

func (p *WorkerPool) Enqueue(ctx context.Context, task Task) error {
	taskCtx := context.WithoutCancel(ctx)

	select {
	case p.queue <- func() { task(p.poolCtx, taskCtx) }: // context.WithDeadline(context.Background(), p.deadline)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *WorkerPool) Stop(timeout time.Duration) error {
	if !p.isStopping.CompareAndSwap(false, true) {
		return nil
	}

	close(p.stop)

	// p.cancelTask()

	// p.deadline := time.Now() + timeout

	done := make(chan struct{})

	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return errors.New("stop timeout")
	}
}
