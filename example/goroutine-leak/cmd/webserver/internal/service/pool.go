package service

import (
	"context"
	"sync"
)

type WorkerPool struct {
	queue chan func(ctx context.Context)
	count int
	stop  chan struct{}
	wg    sync.WaitGroup
}

func NewWorkerPool(count int) *WorkerPool {
	pool := &WorkerPool{
		queue: make(chan func(ctx context.Context), count),
		count: count,
		stop:  make(chan struct{}),
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
			if task != nil {
				task(context.Background())
			}
		}
	}
}

func (p *WorkerPool) Enqueue(ctx context.Context, task func(ctx context.Context)) error {
	select {
	case p.queue <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *WorkerPool) Stop(ctx context.Context) {
	close(p.stop)
	p.wg.Wait()
	close(p.queue)
}
