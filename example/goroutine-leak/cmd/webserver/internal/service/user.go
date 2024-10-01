package service

import (
	"context"
	"errors"
	"sync/atomic"
)

type Task func(ctx context.Context)

type Pool struct {
	tasks  chan func()
	stop   atomic.Bool
	stopCh chan struct{}
}

func NewPool(count int) *Pool {
	pool := &Pool{
		tasks:  make(chan func(), count),
		stop:   atomic.Bool{},
		stopCh: make(chan struct{}),
	}

	for i := 0; i < count; i++ {
		go func() {
			for {
				select {
				case task := <-pool.tasks:
					task()
				case <-pool.stopCh:
					return
				}
			}
		}()
	}

	return pool
}

func (p *Pool) Enqueue(ctx context.Context, task Task) error {
	if p.stop.Load() {
		return errors.New("Pool is stopping") // ErrPoolIsStopping
	}

	select {
	case p.tasks <- func() { task(ctx) }:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *Pool) Stop() error {
	p.stop.Store(true)

	close(p.stopCh)

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
	pool      *Pool
}

func NewUserService(repo UserRepository, analytics Analytics) *UserService {
	return &UserService{
		repo:      repo,
		analytics: analytics,
		pool:      NewPool(100),
	}
}

func (s *UserService) Create(ctx context.Context, name, email string) error {
	// create user in the database
	userID, _ := s.repo.Create(ctx, name, email)

	_ = s.pool.Enqueue(ctx, func(ctx context.Context) {
		s.analytics.Send(ctx, "user created", userID)
	})

	return nil
}
