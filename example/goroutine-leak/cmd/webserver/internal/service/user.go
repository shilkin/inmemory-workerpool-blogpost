package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

//go:generate mockgen -source=$GOFILE -destination=internal/mock/$GOFILE

type WorkerPool struct {
	tasks  chan func()
	stop   atomic.Bool
	stopCh chan struct{}
	wg     sync.WaitGroup

	poolCtx context.Context
	cancel  context.CancelFunc
}

func NewPool(count int) *WorkerPool {
	poolCtx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		tasks:   make(chan func(), count),
		stop:    atomic.Bool{},
		stopCh:  make(chan struct{}),
		poolCtx: poolCtx,
		cancel:  cancel,
	}

	pool.wg.Add(count) // counter+=100

	for i := 0; i < count; i++ {
		go func() {
			defer pool.wg.Done() // counter--

			for {
				select {
				case task := <-pool.tasks:
					task() // blocking call
				case <-pool.stopCh:
					return
				}
			}
		}()
	}

	return pool
}

func mergeContext(poolCtx, taskCtx context.Context) context.Context {
	taskCtx, cancel := context.WithCancel(taskCtx)

	go func() {
		select {
		case <-poolCtx.Done():
			cancel()
		case <-taskCtx.Done():
			cancel()
		}
	}()

	return taskCtx
}

func (p *WorkerPool) Enqueue(ctx context.Context, task func(poolCtx, taskCtx context.Context)) error {
	if p.stop.Load() {
		return errors.New("Pool is stopping") // ErrPoolIsStopping
	}

	taskCtx := context.WithoutCancel(ctx) // <-

	// (2) task(/*give it a time to wrap up*/, taskCtx) // <-

	select {
	case p.tasks <- func() { task(p.poolCtx, taskCtx) }:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *WorkerPool) Stop(timeout time.Duration) error {
	if !p.stop.CompareAndSwap(false, true) {
		return nil
	}

	close(p.stopCh) // send a signal to stop

	p.cancel() // <-poolCtx.Done()

	doneCh := make(chan struct{})

	go func() {
		p.wg.Wait()   // hangs
		close(doneCh) // send a signal that we are done
	}()

	select {
	case <-time.After(timeout): // timeout
		return errors.New("stop timeout") // ErrStopTimout
	case <-doneCh:
	}

	return nil
}

//------------------

type UserRepository interface {
	Create(ctx context.Context, name, email string) (string, error)
}

type Analytics interface {
	Send(ctx context.Context, message string, args ...string)
}

type UserService struct {
	repo      UserRepository
	analytics Analytics
	pool      Pool
}

type Pool interface {
	Enqueue(ctx context.Context, task func(poolCtx, taskCtx context.Context)) error
}

func NewUserService(repo UserRepository, analytics Analytics, pool Pool) *UserService {
	return &UserService{
		repo:      repo,
		analytics: analytics,
		pool:      pool,
	}
}

func (s *UserService) Create(ctx context.Context, name, email string) error {
	// create user in the database
	userID, err := s.repo.Create(ctx, name, email)
	if err != nil {
		return fmt.Errorf("repo create: %w", err)
	}

	err = s.pool.Enqueue(ctx, func(_, ctx context.Context) {
		s.analytics.Send(ctx, "user created", userID)
	})
	if err != nil {
		return fmt.Errorf("pool enqueue: %w", err)
	}

	return nil
}
