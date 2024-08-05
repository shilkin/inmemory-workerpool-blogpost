package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

// 1. Graceful shutdown
// 2. *Pass singal about pool stopping to task handlers
// 3. General purpose pool
type Task func(ctx context.Context)

type Pool interface {
	Enqueue(ctx context.Context, task Task) error
	Stop()
}

type PoolService struct {
	wg      sync.WaitGroup
	ch      chan TaskMessage
	running int32
}

type TaskMessage struct {
	task Task
	ctx  context.Context
}

func (ps *PoolService) Enqueue(ctx context.Context, task Task) error {
	if atomic.LoadInt32(&ps.running) != 1 {
		return errors.New("Worker are stopped")
	}

	if ps.ch == nil {
		return errors.New("channel is nil")
	}

	ps.ch <- TaskMessage{ctx: ctx, task: task}
	return nil
}

func (ps *PoolService) Stop() {
	atomic.StoreInt32(&ps.running, 0)
	close(ps.ch)
	ps.wg.Wait()
}

func (ps *PoolService) startWorker(ch chan TaskMessage) {
	defer ps.wg.Done()
	for {
		taskMessage, ok := <-ch
		if !ok {
			return
		}
		taskMessage.task(taskMessage.ctx)
	}
}

func NewPoolService(maxWorkers int) *PoolService {
	ps := PoolService{
		running: 1,
		ch:      make(chan TaskMessage, 10000),
	}

	for i := 0; i < maxWorkers; i++ {
		ps.wg.Add(1)
		go ps.startWorker(ps.ch)
	}

	return &ps
}

type UserRepository interface {
	Create(ctx context.Context, name, email string) (string, error)
}

type Analytics interface {
	Send(ctx context.Context, message string, args ...string)
}

type UserService struct {
	repo        UserRepository
	analytics   Analytics
	poolC       chan struct{}
	poolService *PoolService
}

func NewUserService(repo UserRepository, analytics Analytics) *UserService {
	poolService := NewPoolService(1000)
	return &UserService{
		repo:        repo,
		analytics:   analytics,
		poolC:       make(chan struct{}, 10),
		poolService: poolService,
	}
}

func (s *UserService) Create(ctx context.Context, name, email string) error {
	// create user in the database
	userID, _ := s.repo.Create(ctx, name, email)
	// s.analytics.Send(ctx, "user created", userID)
	s.poolService.Enqueue(ctx, func(ctx context.Context) { s.analytics.Send(context.WithoutCancel(ctx), "user created", userID) })

	// select {
	// case s.poolC <- struct{}{}: // acquire worker
	// case <-ctx.Done():
	// 	return ctx.Err() // out of resources
	// default:
	// }

	// go func() {
	// 	defer func() { <-s.poolC }() // release worker

	// 	s.analytics.Send(context.WithoutCancel(ctx), "user created", userID)
	// }()

	// go s.analytics.Send(context.WithoutCancel(ctx), "user created", userID)

	return nil
}
