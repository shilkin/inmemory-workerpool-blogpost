package service

import (
	"context"
)

type Task func(ctx context.Context)

type TaskEnqueuer interface {
	Enqueue(ctx context.Context, task Task) error
}

type PoolStopper interface {
	Stop( /*timeout*/) error
}

type WorkerPool interface {
	TaskEnqueuer
	PoolStopper
}

type UserRepository interface {
	Create(ctx context.Context, name, email string) (string, error)
}

type Analytics interface {
	Send(ctx context.Context, message string, args ...string)
}

type UserService struct {
	repo      UserRepository
	analytics Analytics
	pool      TaskEnqueuer
}

func NewUserService(repo UserRepository, analytics Analytics) *UserService {
	return &UserService{
		repo:      repo,
		analytics: analytics,
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
