package service

import (
	"context"
	"fmt"
)

//go:generate mockgen -source=$GOFILE -destination=internal/mock/$GOFILE

type Pool interface {
	Enqueue(ctx context.Context, task func(ctx context.Context)) error
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
	pool      Pool
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

	err = s.pool.Enqueue(ctx,
		func(ctx context.Context) {
			s.analytics.Send(context.WithoutCancel(ctx), "user created", userID)
		},
	)
	if err != nil {
		return fmt.Errorf("pool enqueue: %w", err)
	}

	return nil
}
