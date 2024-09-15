package service

import (
	"context"
	"fmt"
	"time"
)

// type WorkerPool interface {
// 	// Enqueue enqueues new task
// 	Enqueue(ctx context.Context, task func(ctx context.Context)) error

// 	// Stop gracefully stops the pool
// 	// Stop(/*ctx?*/)
// }

type UserRepository interface {
	Create(ctx context.Context, name, email string) (string, error)
}

type Analytics interface {
	Send(ctx context.Context, message string, args ...string)
}

type UserService struct {
	repo      UserRepository
	analytics Analytics
	sema      chan struct{}
	pool      WorkerPool
}

func NewUserService(repo UserRepository, analytics Analytics) *UserService {
	return &UserService{
		repo:      repo,
		analytics: analytics,
		sema:      make(chan struct{}, 100),
	}
}

// https://pkg.go.dev/golang.org/x/sync/semaphore
func (s *UserService) Create(ctx context.Context, name, email string) error {
	// create user in the database
	userID, _ := s.repo.Create(ctx, name, email)

	err := s.pool.Enqueue(ctx, func(ctx context.Context) {
		s.analytics.Send(ctx, "user created", userID)
	})
	if err != nil {
		return fmt.Errorf("enqueue task: %w", err)
	}

	return nil

	select {
	case s.sema <- struct{}{}: // acquire
		defer func() { <-s.sema }() // release
	case <-ctx.Done():
		return nil // + log
		// opt 0: ctx.Err() -- context cancelled / deadline exceeded
		// opt 1: return ErrNoWorkers (+ctx.Err())
	}

	ctx = context.WithoutCancel(ctx)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Microsecond)
	defer cancel()

	// send analytics event synchronously
	// which may cause goroutine and memory leak
	go s.analytics.Send(ctx, "user created", userID)

	return nil
}
